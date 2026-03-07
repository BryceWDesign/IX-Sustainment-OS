package domain

import "time"

type CaseState string

const (
	CaseStateNew              CaseState = "new"
	CaseStateTriage           CaseState = "triage"
	CaseStateAwaitingData     CaseState = "awaiting-data"
	CaseStateAwaitingParts    CaseState = "awaiting-parts"
	CaseStateAwaitingApproval CaseState = "awaiting-approval"
	CaseStateActionable       CaseState = "actionable"
	CaseStateDeferred         CaseState = "deferred"
	CaseStateResolved         CaseState = "resolved"
	CaseStateClosed           CaseState = "closed"
)

type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

type Priority string

const (
	PriorityRoutine  Priority = "routine"
	PriorityElevated Priority = "elevated"
	PriorityUrgent   Priority = "urgent"
	PriorityCritical Priority = "critical"
)

type MissionEffect string

const (
	MissionEffectNone     MissionEffect = "none"
	MissionEffectMinor    MissionEffect = "minor"
	MissionEffectModerate MissionEffect = "moderate"
	MissionEffectMajor    MissionEffect = "major"
	MissionEffectCritical MissionEffect = "critical"
)

type BlockerCategory string

const (
	BlockerCategoryData        BlockerCategory = "data"
	BlockerCategoryProcedure   BlockerCategory = "procedure"
	BlockerCategoryEntitlement BlockerCategory = "entitlement"
	BlockerCategoryParts       BlockerCategory = "parts"
	BlockerCategoryTooling     BlockerCategory = "tooling"
	BlockerCategoryApproval    BlockerCategory = "approval"
	BlockerCategoryCapacity    BlockerCategory = "capacity"
	BlockerCategoryPolicy      BlockerCategory = "policy"
)

type RecommendationType string

const (
	RecommendationTypeLikelyFaultFamily   RecommendationType = "likely-fault-family"
	RecommendationTypeNextEvidence        RecommendationType = "next-evidence"
	RecommendationTypeLikelyBlockerCause  RecommendationType = "likely-blocker-cause"
	RecommendationTypeProcedureSuggestion RecommendationType = "procedure-suggestion"
	RecommendationTypeQueuePriorityHint   RecommendationType = "queue-priority-hint"
	RecommendationTypeSimilarCaseRetrieval RecommendationType = "similar-case-retrieval"
)

type RecommendationStatus string

const (
	RecommendationStatusPendingReview RecommendationStatus = "pending-review"
	RecommendationStatusAccepted      RecommendationStatus = "accepted"
	RecommendationStatusRejected      RecommendationStatus = "rejected"
	RecommendationStatusOverridden    RecommendationStatus = "overridden"
	RecommendationStatusDeferred      RecommendationStatus = "deferred"
)

type ConfidenceLabel string

const (
	ConfidenceLabelLow       ConfidenceLabel = "low"
	ConfidenceLabelMedium    ConfidenceLabel = "medium"
	ConfidenceLabelHigh      ConfidenceLabel = "high"
	ConfidenceLabelUncertain ConfidenceLabel = "uncertain"
)

type ApprovalDisposition string

const (
	ApprovalDispositionPending         ApprovalDisposition = "pending"
	ApprovalDispositionApproved        ApprovalDisposition = "approved"
	ApprovalDispositionRejected        ApprovalDisposition = "rejected"
	ApprovalDispositionReturnedForInfo ApprovalDisposition = "returned-for-info"
)

type AccessState string

const (
	AccessStateAccessible AccessState = "accessible"
	AccessStateRestricted AccessState = "restricted"
	AccessStateUnknown    AccessState = "unknown"
	AccessStateOutdated   AccessState = "outdated"
	AccessStateConflict   AccessState = "conflict"
)

type AvailabilityState string

const (
	AvailabilityStateAvailable   AvailabilityState = "available"
	AvailabilityStateConstrained AvailabilityState = "constrained"
	AvailabilityStateUnavailable AvailabilityState = "unavailable"
	AvailabilityStateBackordered AvailabilityState = "backordered"
	AvailabilityStateUnknown     AvailabilityState = "unknown"
)

type ActorRef struct {
	ActorID     string `json:"actor_id"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type Asset struct {
	AssetID          string        `json:"asset_id"`
	AssetType        string        `json:"asset_type"`
	TailOrSerial     string        `json:"tail_or_serial"`
	UnitOrLocation   string        `json:"unit_or_location"`
	Status           string        `json:"status"`
	MissionRelevance MissionEffect `json:"mission_relevance"`
}

type Blocker struct {
	Category  BlockerCategory `json:"category"`
	Summary   string          `json:"summary"`
	IsPrimary bool            `json:"is_primary"`
	CreatedAt time.Time       `json:"created_at,omitempty"`
	CreatedBy *ActorRef       `json:"created_by,omitempty"`
}

type AttachmentRef struct {
	AttachmentID string    `json:"attachment_id"`
	Filename     string    `json:"filename"`
	MediaType    string    `json:"media_type,omitempty"`
	UploadedAt   time.Time `json:"uploaded_at,omitempty"`
	UploadedBy   *ActorRef `json:"uploaded_by,omitempty"`
}

type FaultEvent struct {
	FaultEventID string    `json:"fault_event_id"`
	CaseID       string    `json:"case_id"`
	Type         string    `json:"type"`
	Source       string    `json:"source"`
	ObservedAt   time.Time `json:"observed_at"`
	Notes        string    `json:"notes"`
}

type Case struct {
	CaseID            string           `json:"case_id"`
	Title             string           `json:"title"`
	Description       string           `json:"description"`
	Severity          Severity         `json:"severity"`
	Priority          Priority         `json:"priority"`
	State             CaseState        `json:"state"`
	AssetID           string           `json:"asset_id"`
	MissionEffect     MissionEffect    `json:"mission_effect"`
	CreatedAt         time.Time        `json:"created_at"`
	CreatedBy         ActorRef         `json:"created_by"`
	UpdatedAt         *time.Time       `json:"updated_at,omitempty"`
	UpdatedBy         *ActorRef        `json:"updated_by,omitempty"`
	PrimaryBlocker    BlockerCategory  `json:"primary_blocker"`
	BlockerList       []Blocker        `json:"blocker_list"`
	ReportedCondition string           `json:"reported_condition,omitempty"`
	SubsystemArea     string           `json:"subsystem_area,omitempty"`
	UrgencyNote       string           `json:"urgency_note,omitempty"`
	ApprovalRequired  bool             `json:"approval_required,omitempty"`
	HasRecommendation bool             `json:"has_recommendation,omitempty"`
	Tags              []string         `json:"tags,omitempty"`
	Attachments       []AttachmentRef  `json:"attachments,omitempty"`
	FaultEvents       []FaultEvent     `json:"fault_events,omitempty"`
	ClosedAt          *time.Time       `json:"closed_at,omitempty"`
	ClosedBy          *ActorRef        `json:"closed_by,omitempty"`
}

type ProcedureRef struct {
	ProcedureRefID string      `json:"procedure_ref_id"`
	Title          string      `json:"title"`
	ReferenceCode  string      `json:"reference_code"`
	Revision       string      `json:"revision"`
	Applicability  string      `json:"applicability"`
	AccessState    AccessState `json:"access_state"`
	RestrictedReason string    `json:"restricted_reason,omitempty"`
}

type PartConstraint struct {
	PartConstraintID string            `json:"part_constraint_id"`
	CaseID           string            `json:"case_id"`
	PartNumber       string            `json:"part_number"`
	Nomenclature     string            `json:"nomenclature"`
	AvailabilityState AvailabilityState `json:"availability_state"`
	ETAText          string            `json:"eta_text,omitempty"`
	ReadinessImpact  string            `json:"readiness_impact"`
	AlternatePath    string            `json:"alternate_path,omitempty"`
}

type Recommendation struct {
	RecommendationID    string               `json:"recommendation_id"`
	CaseID              string               `json:"case_id"`
	Type                RecommendationType   `json:"type"`
	Summary             string               `json:"summary"`
	Rationale           string               `json:"rationale"`
	ConfidenceLabel     ConfidenceLabel      `json:"confidence_label"`
	ApprovalRequired    bool                 `json:"approval_required"`
	Status              RecommendationStatus `json:"status"`
	GeneratedAt         time.Time            `json:"generated_at"`
	GeneratedBy         string               `json:"generated_by,omitempty"`
	InputsUsed          []string             `json:"inputs_used,omitempty"`
	PolicyContext       string               `json:"policy_context,omitempty"`
	ReviewedAt          *time.Time           `json:"reviewed_at,omitempty"`
	ReviewedBy          *ActorRef            `json:"reviewed_by,omitempty"`
	ReviewDispositionReason string           `json:"review_disposition_reason,omitempty"`
	OverrideSummary     string               `json:"override_summary,omitempty"`
	ApprovalID          string               `json:"approval_id,omitempty"`
	ExpiresAt           *time.Time           `json:"expires_at,omitempty"`
	Stale               bool                 `json:"stale,omitempty"`
	Tags                []string             `json:"tags,omitempty"`
}

type Approval struct {
	ApprovalID        string              `json:"approval_id"`
	RelatedObjectType string              `json:"related_object_type"`
	RelatedObjectID   string              `json:"related_object_id"`
	CaseID            string              `json:"case_id,omitempty"`
	RequestedAction   string              `json:"requested_action"`
	RequestReason     string              `json:"request_reason,omitempty"`
	Requester         ActorRef            `json:"requester"`
	Approver          *ActorRef           `json:"approver,omitempty"`
	ApproverRole      string              `json:"approver_role"`
	Disposition       ApprovalDisposition `json:"disposition"`
	RequestedAt       time.Time           `json:"requested_at"`
	DueAt             *time.Time          `json:"due_at,omitempty"`
	DecidedAt         *time.Time          `json:"decided_at,omitempty"`
	DecisionReason    string              `json:"decision_reason,omitempty"`
	StatusSummary     string              `json:"status_summary,omitempty"`
	PolicyContext     string              `json:"policy_context,omitempty"`
	Stale             bool                `json:"stale,omitempty"`
}

type ObjectRef struct {
	ObjectType string `json:"object_type"`
	ObjectID   string `json:"object_id"`
	Summary    string `json:"summary,omitempty"`
}

type ChangeItem struct {
	Field   string `json:"field"`
	Before  any    `json:"before,omitempty"`
	After   any    `json:"after,omitempty"`
	Summary string `json:"summary,omitempty"`
}

type RecommendationContext struct {
	RecommendationID string               `json:"recommendation_id,omitempty"`
	RecommendationType RecommendationType `json:"recommendation_type,omitempty"`
	StatusBefore     RecommendationStatus `json:"status_before,omitempty"`
	StatusAfter      RecommendationStatus `json:"status_after,omitempty"`
	ConfidenceLabel  ConfidenceLabel      `json:"confidence_label,omitempty"`
	ApprovalRequired bool                 `json:"approval_required,omitempty"`
}

type ApprovalContext struct {
	ApprovalID         string              `json:"approval_id,omitempty"`
	RequestedAction    string              `json:"requested_action,omitempty"`
	DispositionBefore  ApprovalDisposition `json:"disposition_before,omitempty"`
	DispositionAfter   ApprovalDisposition `json:"disposition_after,omitempty"`
	ApproverRole       string              `json:"approver_role,omitempty"`
}

type EntitlementContext struct {
	AccessState  AccessState `json:"access_state,omitempty"`
	ResourceType string      `json:"resource_type,omitempty"`
	ResourceID   string      `json:"resource_id,omitempty"`
	Reason       string      `json:"reason,omitempty"`
}

type IntegrityContext struct {
	TraceID       string `json:"trace_id,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`
	SourceSystem  string `json:"source_system,omitempty"`
	Ingested      bool   `json:"ingested,omitempty"`
}

type EvidenceEvent struct {
	EventID               string                 `json:"event_id"`
	ObjectType            string                 `json:"object_type"`
	ObjectID              string                 `json:"object_id"`
	Action                string                 `json:"action"`
	Actor                 ActorRef               `json:"actor"`
	OccurredAt            time.Time              `json:"occurred_at"`
	Summary               string                 `json:"summary"`
	PolicyContext         string                 `json:"policy_context,omitempty"`
	BeforeState           string                 `json:"before_state,omitempty"`
	AfterState            string                 `json:"after_state,omitempty"`
	RelatedObjects        []ObjectRef            `json:"related_objects,omitempty"`
	RecommendationContext *RecommendationContext `json:"recommendation_context,omitempty"`
	ApprovalContext       *ApprovalContext       `json:"approval_context,omitempty"`
	EntitlementContext    *EntitlementContext    `json:"entitlement_context,omitempty"`
	Changes               []ChangeItem           `json:"changes,omitempty"`
	Metadata              map[string]any         `json:"metadata,omitempty"`
	Integrity             *IntegrityContext      `json:"integrity,omitempty"`
}
