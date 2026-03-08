package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/BryceWDesign/IX-Sustainment-OS/internal/domain"
)

const (
	defaultAddr         = ":8080"
	defaultReadTimeout  = 10 * time.Second
	defaultWriteTimeout = 15 * time.Second
	defaultIdleTimeout  = 60 * time.Second
	shutdownTimeout     = 10 * time.Second
)

type appInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	Status      string `json:"status"`
	TimeUTC     string `json:"time_utc"`
}

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	TimeUTC string `json:"time_utc"`
}

type errorResponse struct {
	Error   string         `json:"error"`
	Message string         `json:"message,omitempty"`
	Details map[string]any `json:"details,omitempty"`
}

type caseListResponse struct {
	Items  []domain.Case `json:"items"`
	Total  int           `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

type caseDetailResponse struct {
	Case            domain.Case             `json:"case"`
	Asset           domain.Asset            `json:"asset"`
	FaultEvents     []domain.FaultEvent     `json:"fault_events,omitempty"`
	Procedures      []domain.ProcedureRef   `json:"procedures,omitempty"`
	PartConstraints []domain.PartConstraint `json:"part_constraints,omitempty"`
	Recommendations []domain.Recommendation `json:"recommendations,omitempty"`
	Approvals       []domain.Approval       `json:"approvals,omitempty"`
}

type recommendationListResponse struct {
	Items []domain.Recommendation `json:"items"`
}

type approvalListResponse struct {
	Items  []domain.Approval `json:"items"`
	Total  int               `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

type evidenceEventListResponse struct {
	Items  []domain.EvidenceEvent `json:"items"`
	Total  int                    `json:"total"`
	Limit  int                    `json:"limit"`
	Offset int                    `json:"offset"`
}

type caseRecord struct {
	Case            domain.Case
	Asset           domain.Asset
	FaultEvents     []domain.FaultEvent
	Procedures      []domain.ProcedureRef
	PartConstraints []domain.PartConstraint
	Recommendations []domain.Recommendation
	Approvals       []domain.Approval
	Evidence        []domain.EvidenceEvent
}

type demoStore struct {
	caseOrder []string
	cases     map[string]caseRecord
}

func main() {
	logger := log.New(os.Stdout, "[ix-sustainment-os] ", log.LstdFlags|log.LUTC|log.Lmsgprefix)

	addr := envOrDefault("IX_SUSTAINMENT_OS_ADDR", defaultAddr)
	version := envOrDefault("IX_SUSTAINMENT_OS_VERSION", "dev")
	environment := envOrDefault("IX_SUSTAINMENT_OS_ENV", "local")
	store := seedDemoStore()

	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler(version, environment))
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/readyz", readinessHandler)
	mux.HandleFunc("/version", versionHandler(version, environment))
	mux.HandleFunc("/cases", store.handleCases)
	mux.HandleFunc("/cases/", store.handleCaseSubroutes)
	mux.HandleFunc("/approvals", store.handleApprovals)

	server := &http.Server{
		Addr:         addr,
		Handler:      requestLoggingMiddleware(logger, mux),
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
	}

	logger.Printf("starting server on %s (env=%s version=%s)", addr, environment, version)

	serverErrCh := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
			return
		}
		serverErrCh <- nil
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stopCh:
		logger.Printf("shutdown signal received: %s", sig.String())

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Printf("graceful shutdown failed: %v", err)
			if closeErr := server.Close(); closeErr != nil {
				logger.Printf("forced close failed: %v", closeErr)
			}
			os.Exit(1)
		}

		logger.Printf("server stopped cleanly")
	case err := <-serverErrCh:
		if err != nil {
			logger.Printf("server failed: %v", err)
			os.Exit(1)
		}
	}
}

func rootHandler(version, environment string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload := appInfo{
			Name:        "IX Sustainment OS",
			Description: "CUI-conscious sustainment operating layer for maintenance, readiness, parts bottlenecks, technical-data access, and auditable AI-assisted decision support.",
			Version:     version,
			Environment: environment,
			Status:      "running",
			TimeUTC:     time.Now().UTC().Format(time.RFC3339),
		}

		writeJSON(w, http.StatusOK, payload)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	payload := healthResponse{
		Status:  "ok",
		Service: "ix-sustainment-os",
		TimeUTC: time.Now().UTC().Format(time.RFC3339),
	}

	writeJSON(w, http.StatusOK, payload)
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	payload := healthResponse{
		Status:  "ready",
		Service: "ix-sustainment-os",
		TimeUTC: time.Now().UTC().Format(time.RFC3339),
	}

	writeJSON(w, http.StatusOK, payload)
}

func versionHandler(version, environment string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload := map[string]string{
			"service":     "ix-sustainment-os",
			"version":     version,
			"environment": environment,
			"time_utc":    time.Now().UTC().Format(time.RFC3339),
		}

		writeJSON(w, http.StatusOK, payload)
	}
}

func (s *demoStore) handleCases(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET is supported for /cases in the demo backend.", nil)
		return
	}

	limit := parsePositiveInt(r.URL.Query().Get("limit"), 50)
	offset := parseNonNegativeInt(r.URL.Query().Get("offset"), 0)

	stateFilter := strings.TrimSpace(r.URL.Query().Get("state"))
	severityFilter := strings.TrimSpace(r.URL.Query().Get("severity"))
	primaryBlockerFilter := strings.TrimSpace(r.URL.Query().Get("primary_blocker"))
	assetIDFilter := strings.TrimSpace(r.URL.Query().Get("asset_id"))
	approvalFlagFilter := parseOptionalBool(r.URL.Query().Get("requires_approval"))
	recommendationFlagFilter := parseOptionalBool(r.URL.Query().Get("has_recommendation"))

	items := make([]domain.Case, 0, len(s.caseOrder))
	for _, id := range s.caseOrder {
		record := s.cases[id]
		if stateFilter != "" && string(record.Case.State) != stateFilter {
			continue
		}
		if severityFilter != "" && string(record.Case.Severity) != severityFilter {
			continue
		}
		if primaryBlockerFilter != "" && string(record.Case.PrimaryBlocker) != primaryBlockerFilter {
			continue
		}
		if assetIDFilter != "" && record.Case.AssetID != assetIDFilter {
			continue
		}
		if approvalFlagFilter != nil && record.Case.ApprovalRequired != *approvalFlagFilter {
			continue
		}
		if recommendationFlagFilter != nil && record.Case.HasRecommendation != *recommendationFlagFilter {
			continue
		}

		items = append(items, record.Case)
	}

	total := len(items)
	items = slicePage(items, offset, limit)

	writeJSON(w, http.StatusOK, caseListResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func (s *demoStore) handleCaseSubroutes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET is supported for /cases/* demo routes.", nil)
		return
	}

	path := strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/"), "/cases/")
	if path == "" {
		writeErrorJSON(w, http.StatusNotFound, "not_found", "Case route not found.", nil)
		return
	}

	parts := strings.Split(path, "/")
	caseID := parts[0]

	record, ok := s.cases[caseID]
	if !ok {
		writeErrorJSON(w, http.StatusNotFound, "case_not_found", "Requested case was not found.", map[string]any{
			"case_id": caseID,
		})
		return
	}

	if len(parts) == 1 {
		writeJSON(w, http.StatusOK, caseDetailResponse{
			Case:            record.Case,
			Asset:           record.Asset,
			FaultEvents:     record.FaultEvents,
			Procedures:      record.Procedures,
			PartConstraints: record.PartConstraints,
			Recommendations: record.Recommendations,
			Approvals:       record.Approvals,
		})
		return
	}

	switch parts[1] {
	case "evidence":
		s.handleCaseEvidence(w, r, record)
		return
	case "recommendations":
		s.handleCaseRecommendations(w, r, record)
		return
	default:
		writeErrorJSON(w, http.StatusNotFound, "not_found", "Case subroute not found.", map[string]any{
			"path": r.URL.Path,
		})
		return
	}
}

func (s *demoStore) handleCaseRecommendations(w http.ResponseWriter, r *http.Request, record caseRecord) {
	writeJSON(w, http.StatusOK, recommendationListResponse{
		Items: record.Recommendations,
	})
}

func (s *demoStore) handleCaseEvidence(w http.ResponseWriter, r *http.Request, record caseRecord) {
	limit := parsePositiveInt(r.URL.Query().Get("limit"), 50)
	offset := parseNonNegativeInt(r.URL.Query().Get("offset"), 0)
	actionFilter := strings.TrimSpace(r.URL.Query().Get("action"))
	actorFilter := strings.TrimSpace(r.URL.Query().Get("actor"))

	items := make([]domain.EvidenceEvent, 0, len(record.Evidence))
	for _, event := range record.Evidence {
		if actionFilter != "" && event.Action != actionFilter {
			continue
		}
		if actorFilter != "" && event.Actor.ActorID != actorFilter && event.Actor.DisplayName != actorFilter {
			continue
		}
		items = append(items, event)
	}

	total := len(items)
	items = slicePage(items, offset, limit)

	writeJSON(w, http.StatusOK, evidenceEventListResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func (s *demoStore) handleApprovals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorJSON(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET is supported for /approvals in the demo backend.", nil)
		return
	}

	statusFilter := strings.TrimSpace(r.URL.Query().Get("status"))
	caseIDFilter := strings.TrimSpace(r.URL.Query().Get("case_id"))
	approverRoleFilter := strings.TrimSpace(r.URL.Query().Get("approver_role"))
	limit := parsePositiveInt(r.URL.Query().Get("limit"), 50)
	offset := parseNonNegativeInt(r.URL.Query().Get("offset"), 0)

	var approvals []domain.Approval
	for _, id := range s.caseOrder {
		record := s.cases[id]
		for _, approval := range record.Approvals {
			if statusFilter != "" && string(approval.Disposition) != statusFilter {
				continue
			}
			if caseIDFilter != "" && approval.CaseID != caseIDFilter {
				continue
			}
			if approverRoleFilter != "" && approval.ApproverRole != approverRoleFilter {
				continue
			}
			approvals = append(approvals, approval)
		}
	}

	total := len(approvals)
	approvals = slicePage(approvals, offset, limit)

	writeJSON(w, http.StatusOK, approvalListResponse{
		Items:  approvals,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func requestLoggingMiddleware(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now().UTC()

		recorder := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		logger.Printf(
			"method=%s path=%s status=%d remote=%s duration_ms=%d",
			r.Method,
			r.URL.Path,
			recorder.statusCode,
			r.RemoteAddr,
			time.Since(started).Milliseconds(),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	_ = encoder.Encode(payload)
}

func writeErrorJSON(w http.ResponseWriter, status int, code string, message string, details map[string]any) {
	writeJSON(w, status, errorResponse{
		Error:   code,
		Message: message,
		Details: details,
	})
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func parsePositiveInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}

	var parsed int
	_, err := fmtSscanfInt(value, &parsed)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}

func parseNonNegativeInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}

	var parsed int
	_, err := fmtSscanfInt(value, &parsed)
	if err != nil || parsed < 0 {
		return fallback
	}

	return parsed
}

func parseOptionalBool(value string) *bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		return nil
	case "true", "1", "yes":
		v := true
		return &v
	case "false", "0", "no":
		v := false
		return &v
	default:
		return nil
	}
}

func fmtSscanfInt(input string, out *int) (int, error) {
	return fmtSscanf(strings.TrimSpace(input), "%d", out)
}

func fmtSscanf(input, format string, a ...any) (int, error) {
	return fmt.Sscanf(input, format, a...)
}

func slicePage[T any](items []T, offset int, limit int) []T {
	if offset >= len(items) {
		return []T{}
	}

	end := offset + limit
	if end > len(items) {
		end = len(items)
	}

	return items[offset:end]
}

func seedDemoStore() *demoStore {
	now := time.Date(2026, time.March, 7, 18, 30, 0, 0, time.UTC)

	productionController := domain.ActorRef{
		ActorID:     "usr-0142",
		DisplayName: "Jordan Hale",
		Role:        "production-controller",
	}

	supervisor := domain.ActorRef{
		ActorID:     "usr-0021",
		DisplayName: "Avery Mercer",
		Role:        "supervisor",
	}

	supplyAnalyst := domain.ActorRef{
		ActorID:     "usr-0188",
		DisplayName: "Dana Voss",
		Role:        "supply-analyst",
	}

	assetA := domain.Asset{
		AssetID:          "AST-00091",
		AssetType:        "support-platform",
		TailOrSerial:     "SN-22914",
		UnitOrLocation:   "Depot-West",
		Status:           "degraded",
		MissionRelevance: domain.MissionEffectMajor,
	}

	assetB := domain.Asset{
		AssetID:          "AST-00107",
		AssetType:        "support-platform",
		TailOrSerial:     "SN-44102",
		UnitOrLocation:   "Depot-West",
		Status:           "inspection-hold",
		MissionRelevance: domain.MissionEffectModerate,
	}

	caseA := domain.Case{
		CaseID:         "CASE-2026-0001",
		Title:          "Hydraulic pressure loss during post-maintenance verification",
		Description:    "Pressure drop observed during verification cycle after routine service. Team suspects seal assembly degradation with substitute-path review required.",
		Severity:       domain.SeverityHigh,
		Priority:       domain.PriorityUrgent,
		State:          domain.CaseStateAwaitingApproval,
		AssetID:        assetA.AssetID,
		MissionEffect:  domain.MissionEffectMajor,
		CreatedAt:      now.Add(-9 * time.Hour),
		CreatedBy:      productionController,
		PrimaryBlocker: domain.BlockerCategoryApproval,
		BlockerList: []domain.Blocker{
			{
				Category:  domain.BlockerCategoryParts,
				Summary:   "Required hydraulic seal assembly is unavailable with substitute path under review.",
				IsPrimary: false,
				CreatedAt: now.Add(-8 * time.Hour),
				CreatedBy: &productionController,
			},
			{
				Category:  domain.BlockerCategoryApproval,
				Summary:   "Substitute-part review requires accountable supervisor sign-off.",
				IsPrimary: true,
				CreatedAt: now.Add(-6 * time.Hour),
				CreatedBy: &productionController,
			},
		},
		ReportedCondition: "Pressure dropped after second verification cycle.",
		SubsystemArea:     "HYD-A",
		UrgencyNote:       "Needed back in rotation quickly for follow-on sustainment queue.",
		ApprovalRequired:  true,
		HasRecommendation: true,
		Tags:              []string{"hydraulics", "verification", "approval-gated"},
	}

	caseB := domain.Case{
		CaseID:         "CASE-2026-0002",
		Title:          "Procedure applicability conflict during actuator inspection",
		Description:    "Inspection team found conflicting revision metadata between local reference packet and current procedure index.",
		Severity:       domain.SeverityMedium,
		Priority:       domain.PriorityElevated,
		State:          domain.CaseStateAwaitingData,
		AssetID:        assetB.AssetID,
		MissionEffect:  domain.MissionEffectModerate,
		CreatedAt:      now.Add(-13 * time.Hour),
		CreatedBy:      productionController,
		PrimaryBlocker: domain.BlockerCategoryProcedure,
		BlockerList: []domain.Blocker{
			{
				Category:  domain.BlockerCategoryProcedure,
				Summary:   "Procedure revision conflict prevents confirmation of the correct inspection path.",
				IsPrimary: true,
				CreatedAt: now.Add(-12 * time.Hour),
				CreatedBy: &productionController,
			},
			{
				Category:  domain.BlockerCategoryEntitlement,
				Summary:   "Current reviewer does not have access to the full restricted reference body.",
				IsPrimary: false,
				CreatedAt: now.Add(-11 * time.Hour),
				CreatedBy: &productionController,
			},
		},
		ReportedCondition: "Conflicting reference packets found at work-center.",
		SubsystemArea:     "ACT-2",
		UrgencyNote:       "Need clarity before inspection lane proceeds.",
		ApprovalRequired:  false,
		HasRecommendation: true,
		Tags:              []string{"procedure", "inspection", "reference-conflict"},
	}

	faultA := domain.FaultEvent{
		FaultEventID: "FE-2026-0030",
		CaseID:       caseA.CaseID,
		Type:         "pressure-loss",
		Source:       "maintainer-entry",
		ObservedAt:   now.Add(-9 * time.Hour),
		Notes:        "Pressure dropped after second verification cycle.",
	}

	faultB := domain.FaultEvent{
		FaultEventID: "FE-2026-0031",
		CaseID:       caseB.CaseID,
		Type:         "procedure-conflict",
		Source:       "planner-entry",
		ObservedAt:   now.Add(-12 * time.Hour),
		Notes:        "Local packet revision does not match current procedure index.",
	}

	procedureA := domain.ProcedureRef{
		ProcedureRefID:  "PROC-1148",
		Title:           "Hydraulic Line Verification Procedure",
		ReferenceCode:   "TM-HYD-1148",
		Revision:        "Rev C",
		Applicability:   "Applies to platform block 2 with subsystem HYD-A.",
		AccessState:     domain.AccessStateRestricted,
		RestrictedReason: "Full procedure view requires elevated entitlement when substitute-path review is active.",
	}

	procedureB := domain.ProcedureRef{
		ProcedureRefID:  "PROC-2207",
		Title:           "Actuator Inspection Procedure",
		ReferenceCode:   "TM-ACT-2207",
		Revision:        "Rev F",
		Applicability:   "Applies to actuator family ACT-2.",
		AccessState:     domain.AccessStateConflict,
		RestrictedReason: "Index indicates Rev F while local packet references Rev E.",
	}

	partConstraintA := domain.PartConstraint{
		PartConstraintID:  "PC-0091",
		CaseID:            caseA.CaseID,
		PartNumber:        "PN-44-771A",
		Nomenclature:      "hydraulic seal assembly",
		AvailabilityState: domain.AvailabilityStateUnavailable,
		ETAText:           "ETA unknown",
		ReadinessImpact:   "Blocks restoration of one mission-relevant asset.",
		AlternatePath:     "Substitute under evaluation, approval required.",
	}

	recommendationA := domain.Recommendation{
		RecommendationID: "REC-2026-0017",
		CaseID:           caseA.CaseID,
		Type:             domain.RecommendationTypeProcedureSuggestion,
		Summary:          "Check revision-aligned seal replacement procedure and verify substitute part eligibility.",
		Rationale:        "Similar prior cases show recurring seal degradation plus revision mismatch risk when substitute-path review is delayed.",
		ConfidenceLabel:  domain.ConfidenceLabelMedium,
		ApprovalRequired: true,
		Status:           domain.RecommendationStatusPendingReview,
		GeneratedAt:      now.Add(-6 * time.Hour),
		GeneratedBy:      "ruleset:v0.1",
		InputsUsed:       []string{"case-summary", "prior-similar-cases", "parts-availability", "procedure-metadata"},
		PolicyContext:    "role=production-controller; approval_required=true; recommendation_class=procedure-suggestion",
	}

	recommendationB := domain.Recommendation{
		RecommendationID: "REC-2026-0018",
		CaseID:           caseB.CaseID,
		Type:             domain.RecommendationTypeNextEvidence,
		Summary:          "Request revision-authoritative procedure packet and compare local reference lineage before continuing inspection.",
		Rationale:        "Current issue is primarily a data and procedure conflict, not a confirmed hardware fault.",
		ConfidenceLabel:  domain.ConfidenceLabelHigh,
		ApprovalRequired: false,
		Status:           domain.RecommendationStatusPendingReview,
		GeneratedAt:      now.Add(-10 * time.Hour),
		GeneratedBy:      "ruleset:v0.1",
		InputsUsed:       []string{"case-summary", "procedure-index", "local-reference-metadata"},
		PolicyContext:    "role=production-controller; approval_required=false; recommendation_class=next-evidence",
	}

	approvalA := domain.Approval{
		ApprovalID:        "APR-2026-0042",
		RelatedObjectType: "recommendation",
		RelatedObjectID:   recommendationA.RecommendationID,
		CaseID:            caseA.CaseID,
		RequestedAction:   "Approve limited review path for substitute-part workflow.",
		RequestReason:     "Restricted workflow branch requires accountable sign-off before advancing the case.",
		Requester:         productionController,
		Approver:          &supervisor,
		ApproverRole:      "supervisor",
		Disposition:       domain.ApprovalDispositionPending,
		RequestedAt:       now.Add(-4 * time.Hour),
		DueAt:             timePointer(now.Add(4 * time.Hour)),
		StatusSummary:     "Supervisor review pending.",
		PolicyContext:     "role=production-controller; transition=awaiting-approval->actionable; approval_required=true",
	}

	evidenceA := []domain.EvidenceEvent{
		{
			EventID:    "EVT-2026-000001",
			ObjectType: "case",
			ObjectID:   caseA.CaseID,
			Action:     "case.created",
			Actor:      productionController,
			OccurredAt: now.Add(-9 * time.Hour),
			Summary:    "Case created with initial sustainment discrepancy details.",
			AfterState: "triage",
			Integrity: &domain.IntegrityContext{
				SourceSystem: "ix-sustainment-os",
			},
		},
		{
			EventID:    "EVT-2026-000002",
			ObjectType: "case",
			ObjectID:   caseA.CaseID,
			Action:     "part_constraint.added",
			Actor:      supplyAnalyst,
			OccurredAt: now.Add(-7 * time.Hour),
			Summary:    "Part constraint added for hydraulic seal assembly.",
			Integrity: &domain.IntegrityContext{
				SourceSystem: "ix-sustainment-os",
			},
		},
		{
			EventID:    "EVT-2026-000003",
			ObjectType: "recommendation",
			ObjectID:   recommendationA.RecommendationID,
			Action:     "recommendation.generated",
			Actor:      productionController,
			OccurredAt: now.Add(-6 * time.Hour),
			Summary:    "Procedure suggestion recommendation generated for the case.",
			Integrity: &domain.IntegrityContext{
				SourceSystem: "ix-sustainment-os",
			},
		},
		{
			EventID:    "EVT-2026-000004",
			ObjectType: "approval",
			ObjectID:   approvalA.ApprovalID,
			Action:     "approval.requested",
			Actor:      productionController,
			OccurredAt: now.Add(-4 * time.Hour),
			Summary:    "Approval requested for substitute-part workflow review path.",
			Integrity: &domain.IntegrityContext{
				SourceSystem: "ix-sustainment-os",
			},
		},
		{
			EventID:     "EVT-2026-000005",
			ObjectType:  "case",
			ObjectID:    caseA.CaseID,
			Action:      "case.state_changed",
			Actor:       productionController,
			OccurredAt:  now.Add(-4 * time.Hour),
			Summary:     "Case moved from triage to awaiting-approval.",
			BeforeState: "triage",
			AfterState:  "awaiting-approval",
			Integrity: &domain.IntegrityContext{
				SourceSystem: "ix-sustainment-os",
			},
		},
	}

	evidenceB := []domain.EvidenceEvent{
		{
			EventID:    "EVT-2026-000006",
			ObjectType: "case",
			ObjectID:   caseB.CaseID,
			Action:     "case.created",
			Actor:      productionController,
			OccurredAt: now.Add(-13 * time.Hour),
			Summary:    "Case created for actuator inspection procedure conflict.",
			AfterState: "triage",
			Integrity: &domain.IntegrityContext{
				SourceSystem: "ix-sustainment-os",
			},
		},
		{
			EventID:    "EVT-2026-000007",
			ObjectType: "case",
			ObjectID:   caseB.CaseID,
			Action:     "entitlement.check_performed",
			Actor:      productionController,
			OccurredAt: now.Add(-11 * time.Hour),
			Summary:    "Entitlement check performed for procedure PROC-2207 with conflict result.",
			Integrity: &domain.IntegrityContext{
				SourceSystem: "ix-sustainment-os",
			},
		},
		{
			EventID:    "EVT-2026-000008",
			ObjectType: "recommendation",
			ObjectID:   recommendationB.RecommendationID,
			Action:     "recommendation.generated",
			Actor:      productionController,
			OccurredAt: now.Add(-10 * time.Hour),
			Summary:    "Next-evidence recommendation generated for revision conflict case.",
			Integrity: &domain.IntegrityContext{
				SourceSystem: "ix-sustainment-os",
			},
		},
		{
			EventID:     "EVT-2026-000009",
			ObjectType:  "case",
			ObjectID:    caseB.CaseID,
			Action:      "case.state_changed",
			Actor:       productionController,
			OccurredAt:  now.Add(-10 * time.Hour),
			Summary:     "Case moved from triage to awaiting-data.",
			BeforeState: "triage",
			AfterState:  "awaiting-data",
			Integrity: &domain.IntegrityContext{
				SourceSystem: "ix-sustainment-os",
			},
		},
	}

	return &demoStore{
		caseOrder: []string{caseA.CaseID, caseB.CaseID},
		cases: map[string]caseRecord{
			caseA.CaseID: {
				Case:            caseA,
				Asset:           assetA,
				FaultEvents:     []domain.FaultEvent{faultA},
				Procedures:      []domain.ProcedureRef{procedureA},
				PartConstraints: []domain.PartConstraint{partConstraintA},
				Recommendations: []domain.Recommendation{recommendationA},
				Approvals:       []domain.Approval{approvalA},
				Evidence:        evidenceA,
			},
			caseB.CaseID: {
				Case:            caseB,
				Asset:           assetB,
				FaultEvents:     []domain.FaultEvent{faultB},
				Procedures:      []domain.ProcedureRef{procedureB},
				Recommendations: []domain.Recommendation{recommendationB},
				Evidence:        evidenceB,
			},
		},
	}
}

func timePointer(v time.Time) *time.Time {
	return &v
}
