package audit

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/BryceWDesign/IX-Sustainment-OS/internal/domain"
)

var eventCounter uint64

type Service struct {
	sourceSystem string
	now          func() time.Time
}

func NewService(sourceSystem string) *Service {
	if strings.TrimSpace(sourceSystem) == "" {
		sourceSystem = "ix-sustainment-os"
	}

	return &Service{
		sourceSystem: sourceSystem,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s *Service) NewCaseCreatedEvent(c domain.Case, actor domain.ActorRef) domain.EvidenceEvent {
	return s.newEvent(
		"case",
		c.CaseID,
		"case.created",
		actor,
		fmt.Sprintf("Case %s created with initial state %s.", c.CaseID, c.State),
		withAfterState(string(c.State)),
		withChange("title", nil, c.Title, "Initial case title recorded."),
		withChange("severity", nil, string(c.Severity), "Initial severity recorded."),
		withChange("priority", nil, string(c.Priority), "Initial priority recorded."),
		withChange("mission_effect", nil, string(c.MissionEffect), "Mission effect recorded."),
	)
}

func (s *Service) NewCaseStateChangedEvent(
	caseID string,
	actor domain.ActorRef,
	from domain.CaseState,
	to domain.CaseState,
	reason string,
	policyContext string,
) domain.EvidenceEvent {
	summary := fmt.Sprintf("Case moved from %s to %s.", from, to)
	if strings.TrimSpace(reason) != "" {
		summary = fmt.Sprintf("%s Reason: %s", summary, reason)
	}

	return s.newEvent(
		"case",
		caseID,
		"case.state_changed",
		actor,
		summary,
		withBeforeState(string(from)),
		withAfterState(string(to)),
		withPolicyContext(policyContext),
		withChange("state", string(from), string(to), "Workflow state updated."),
	)
}

func (s *Service) NewBlockerTaggedEvent(
	caseID string,
	actor domain.ActorRef,
	blocker domain.Blocker,
) domain.EvidenceEvent {
	summary := fmt.Sprintf("Blocker %q tagged as %s.", blocker.Summary, blocker.Category)

	return s.newEvent(
		"case",
		caseID,
		"blocker.tagged",
		actor,
		summary,
		withRelatedObject("blocker", string(blocker.Category), blocker.Summary),
		withChange("primary_blocker", nil, string(blocker.Category), "Blocker category added to active case context."),
	)
}

func (s *Service) NewBlockerClearedEvent(
	caseID string,
	actor domain.ActorRef,
	category domain.BlockerCategory,
	reason string,
) domain.EvidenceEvent {
	summary := fmt.Sprintf("Blocker %s cleared.", category)
	if strings.TrimSpace(reason) != "" {
		summary = fmt.Sprintf("%s Reason: %s", summary, reason)
	}

	return s.newEvent(
		"case",
		caseID,
		"blocker.cleared",
		actor,
		summary,
		withChange("blocker", string(category), nil, "Blocker removed from active case context."),
	)
}

func (s *Service) NewProcedureLinkedEvent(
	caseID string,
	actor domain.ActorRef,
	procedure domain.ProcedureRef,
	policyContext string,
) domain.EvidenceEvent {
	summary := fmt.Sprintf("Procedure %s linked to case.", procedure.ReferenceCode)

	return s.newEvent(
		"case",
		caseID,
		"procedure.linked",
		actor,
		summary,
		withPolicyContext(policyContext),
		withRelatedObject("procedure", procedure.ProcedureRefID, procedure.Title),
		withEntitlementContext(domain.EntitlementContext{
			AccessState:  procedure.AccessState,
			ResourceType: "procedure",
			ResourceID:   procedure.ProcedureRefID,
			Reason:       procedure.RestrictedReason,
		}),
	)
}

func (s *Service) NewProcedureAccessDeniedEvent(
	caseID string,
	actor domain.ActorRef,
	procedure domain.ProcedureRef,
	policyContext string,
) domain.EvidenceEvent {
	summary := fmt.Sprintf("Procedure %s access denied for current workflow context.", procedure.ReferenceCode)

	return s.newEvent(
		"case",
		caseID,
		"procedure.access_denied",
		actor,
		summary,
		withPolicyContext(policyContext),
		withRelatedObject("procedure", procedure.ProcedureRefID, procedure.Title),
		withEntitlementContext(domain.EntitlementContext{
			AccessState:  domain.AccessStateRestricted,
			ResourceType: "procedure",
			ResourceID:   procedure.ProcedureRefID,
			Reason:       procedure.RestrictedReason,
		}),
	)
}

func (s *Service) NewEntitlementCheckPerformedEvent(
	caseID string,
	actor domain.ActorRef,
	resourceType string,
	resourceID string,
	accessState domain.AccessState,
	reason string,
	policyContext string,
) domain.EvidenceEvent {
	summary := fmt.Sprintf("Entitlement check performed for %s %s with result %s.", resourceType, resourceID, accessState)

	return s.newEvent(
		"case",
		caseID,
		"entitlement.check_performed",
		actor,
		summary,
		withPolicyContext(policyContext),
		withEntitlementContext(domain.EntitlementContext{
			AccessState:  accessState,
			ResourceType: resourceType,
			ResourceID:   resourceID,
			Reason:       reason,
		}),
	)
}

func (s *Service) NewPartConstraintAddedEvent(
	caseID string,
	actor domain.ActorRef,
	part domain.PartConstraint,
) domain.EvidenceEvent {
	summary := fmt.Sprintf("Part constraint added for %s (%s).", part.Nomenclature, part.PartNumber)

	return s.newEvent(
		"case",
		caseID,
		"part_constraint.added",
		actor,
		summary,
		withRelatedObject("part-constraint", part.PartConstraintID, part.Nomenclature),
		withChange("availability_state", nil, string(part.AvailabilityState), "Initial part availability state recorded."),
	)
}

func (s *Service) NewSupplyStatusCheckedEvent(
	caseID string,
	actor domain.ActorRef,
	part domain.PartConstraint,
) domain.EvidenceEvent {
	summary := fmt.Sprintf("Supply status checked for %s (%s): %s.", part.Nomenclature, part.PartNumber, part.AvailabilityState)

	return s.newEvent(
		"case",
		caseID,
		"supply_status.checked",
		actor,
		summary,
		withRelatedObject("part-constraint", part.PartConstraintID, part.Nomenclature),
	)
}

func (s *Service) NewReadinessImpactAssessedEvent(
	caseID string,
	actor domain.ActorRef,
	part domain.PartConstraint,
) domain.EvidenceEvent {
	summary := fmt.Sprintf("Readiness impact assessed for %s: %s", part.PartNumber, part.ReadinessImpact)

	return s.newEvent(
		"case",
		caseID,
		"readiness_impact.assessed",
		actor,
		summary,
		withRelatedObject("part-constraint", part.PartConstraintID, part.Nomenclature),
	)
}

func (s *Service) NewRecommendationGeneratedEvent(
	rec domain.Recommendation,
	actor domain.ActorRef,
) domain.EvidenceEvent {
	summary := fmt.Sprintf("Recommendation %s generated for case %s.", rec.Type, rec.CaseID)

	return s.newEvent(
		"recommendation",
		rec.RecommendationID,
		"recommendation.generated",
		actor,
		summary,
		withRelatedObject("case", rec.CaseID, ""),
		withRecommendationContext(domain.RecommendationContext{
			RecommendationID:   rec.RecommendationID,
			RecommendationType: rec.Type,
			StatusAfter:        rec.Status,
			ConfidenceLabel:    rec.ConfidenceLabel,
			ApprovalRequired:   rec.ApprovalRequired,
		}),
	)
}

func (s *Service) NewRecommendationReviewedEvent(
	rec domain.Recommendation,
	actor domain.ActorRef,
	previousStatus domain.RecommendationStatus,
) domain.EvidenceEvent {
	action := recommendationStatusToAction(rec.Status)
	summary := fmt.Sprintf("Recommendation review recorded with final status %s.", rec.Status)
	if strings.TrimSpace(rec.ReviewDispositionReason) != "" {
		summary = fmt.Sprintf("%s Reason: %s", summary, rec.ReviewDispositionReason)
	}

	return s.newEvent(
		"recommendation",
		rec.RecommendationID,
		action,
		actor,
		summary,
		withRelatedObject("case", rec.CaseID, ""),
		withPolicyContext(rec.PolicyContext),
		withRecommendationContext(domain.RecommendationContext{
			RecommendationID:   rec.RecommendationID,
			RecommendationType: rec.Type,
			StatusBefore:       previousStatus,
			StatusAfter:        rec.Status,
			ConfidenceLabel:    rec.ConfidenceLabel,
			ApprovalRequired:   rec.ApprovalRequired,
		}),
		withChange("recommendation.status", string(previousStatus), string(rec.Status), "Recommendation review status updated."),
	)
}

func (s *Service) NewApprovalRequestedEvent(
	approval domain.Approval,
	actor domain.ActorRef,
) domain.EvidenceEvent {
	summary := fmt.Sprintf("Approval requested for action: %s", approval.RequestedAction)

	return s.newEvent(
		"approval",
		approval.ApprovalID,
		"approval.requested",
		actor,
		summary,
		withRelatedObject(approval.RelatedObjectType, approval.RelatedObjectID, ""),
		withRelatedObject("case", approval.CaseID, ""),
		withApprovalContext(domain.ApprovalContext{
			ApprovalID:       approval.ApprovalID,
			RequestedAction:  approval.RequestedAction,
			DispositionAfter: approval.Disposition,
			ApproverRole:     approval.ApproverRole,
		}),
	)
}

func (s *Service) NewApprovalDecisionEvent(
	approval domain.Approval,
	actor domain.ActorRef,
	previousDisposition domain.ApprovalDisposition,
) domain.EvidenceEvent {
	action := approvalDispositionToAction(approval.Disposition)
	summary := fmt.Sprintf("Approval %s recorded as %s.", approval.ApprovalID, approval.Disposition)
	if strings.TrimSpace(approval.DecisionReason) != "" {
		summary = fmt.Sprintf("%s Reason: %s", summary, approval.DecisionReason)
	}

	return s.newEvent(
		"approval",
		approval.ApprovalID,
		action,
		actor,
		summary,
		withRelatedObject(approval.RelatedObjectType, approval.RelatedObjectID, ""),
		withRelatedObject("case", approval.CaseID, ""),
		withPolicyContext(approval.PolicyContext),
		withApprovalContext(domain.ApprovalContext{
			ApprovalID:         approval.ApprovalID,
			RequestedAction:    approval.RequestedAction,
			DispositionBefore:  previousDisposition,
			DispositionAfter:   approval.Disposition,
			ApproverRole:       approval.ApproverRole,
		}),
		withChange("approval.disposition", string(previousDisposition), string(approval.Disposition), "Approval disposition updated."),
	)
}

type eventOption func(*domain.EvidenceEvent)

func withPolicyContext(policyContext string) eventOption {
	return func(e *domain.EvidenceEvent) {
		if strings.TrimSpace(policyContext) != "" {
			e.PolicyContext = policyContext
		}
	}
}

func withBeforeState(state string) eventOption {
	return func(e *domain.EvidenceEvent) {
		if strings.TrimSpace(state) != "" {
			e.BeforeState = state
		}
	}
}

func withAfterState(state string) eventOption {
	return func(e *domain.EvidenceEvent) {
		if strings.TrimSpace(state) != "" {
			e.AfterState = state
		}
	}
}

func withRelatedObject(objectType, objectID, summary string) eventOption {
	return func(e *domain.EvidenceEvent) {
		if strings.TrimSpace(objectType) == "" || strings.TrimSpace(objectID) == "" {
			return
		}

		e.RelatedObjects = append(e.RelatedObjects, domain.ObjectRef{
			ObjectType: objectType,
			ObjectID:   objectID,
			Summary:    summary,
		})
	}
}

func withChange(field string, before any, after any, summary string) eventOption {
	return func(e *domain.EvidenceEvent) {
		if strings.TrimSpace(field) == "" {
			return
		}

		e.Changes = append(e.Changes, domain.ChangeItem{
			Field:   field,
			Before:  before,
			After:   after,
			Summary: summary,
		})
	}
}

func withRecommendationContext(ctx domain.RecommendationContext) eventOption {
	return func(e *domain.EvidenceEvent) {
		e.RecommendationContext = &ctx
	}
}

func withApprovalContext(ctx domain.ApprovalContext) eventOption {
	return func(e *domain.EvidenceEvent) {
		e.ApprovalContext = &ctx
	}
}

func withEntitlementContext(ctx domain.EntitlementContext) eventOption {
	return func(e *domain.EvidenceEvent) {
		e.EntitlementContext = &ctx
	}
}

func (s *Service) newEvent(
	objectType string,
	objectID string,
	action string,
	actor domain.ActorRef,
	summary string,
	opts ...eventOption,
) domain.EvidenceEvent {
	event := domain.EvidenceEvent{
		EventID:    nextEventID(s.now()),
		ObjectType: objectType,
		ObjectID:   objectID,
		Action:     action,
		Actor:      actor,
		OccurredAt: s.now(),
		Summary:    summary,
		Integrity: &domain.IntegrityContext{
			SourceSystem: s.sourceSystem,
		},
	}

	for _, opt := range opts {
		opt(&event)
	}

	return event
}

func nextEventID(now time.Time) string {
	n := atomic.AddUint64(&eventCounter, 1)
	return fmt.Sprintf("EVT-%04d-%06d", now.UTC().Year(), n)
}

func recommendationStatusToAction(status domain.RecommendationStatus) string {
	switch status {
	case domain.RecommendationStatusAccepted:
		return "recommendation.accepted"
	case domain.RecommendationStatusRejected:
		return "recommendation.rejected"
	case domain.RecommendationStatusOverridden:
		return "recommendation.overridden"
	default:
		return "recommendation.reviewed"
	}
}

func approvalDispositionToAction(disposition domain.ApprovalDisposition) string {
	switch disposition {
	case domain.ApprovalDispositionApproved:
		return "approval.approved"
	case domain.ApprovalDispositionRejected:
		return "approval.rejected"
	case domain.ApprovalDispositionReturnedForInfo:
		return "approval.returned_for_info"
	default:
		return "approval.requested"
	}
}
