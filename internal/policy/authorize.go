package policy

import (
	"fmt"

	"github.com/BryceWDesign/IX-Sustainment-OS/internal/domain"
)

type Action string

const (
	ActionCaseCreate           Action = "case.create"
	ActionCaseView             Action = "case.view"
	ActionCaseTransition       Action = "case.transition"
	ActionBlockerUpdate        Action = "blocker.update"
	ActionProcedureLink        Action = "procedure.link"
	ActionPartConstraintAdd    Action = "part-constraint.add"
	ActionRecommendationReview Action = "recommendation.review"
	ActionApprovalRequest      Action = "approval.request"
	ActionApprovalDecide       Action = "approval.decide"
	ActionEvidenceView         Action = "evidence.view"
)

const (
	RoleMaintainer           = "maintainer"
	RoleProductionController = "production-controller"
	RoleSupplyAnalyst        = "supply-analyst"
	RoleSustainmentEngineer  = "sustainment-engineer"
	RoleSupervisor           = "supervisor"
	RolePolicyReviewer       = "policy-reviewer"
	RoleAdministrator        = "administrator"
)

type GuardInput struct {
	ActorRole                   string
	Action                      Action
	CurrentState                domain.CaseState
	TargetState                 domain.CaseState
	RecommendationStatus        domain.RecommendationStatus
	RecommendationNeedsApproval bool
	RestrictedProcedure         bool
	OverrideRequested           bool
}

type Decision struct {
	Allowed          bool   `json:"allowed"`
	ApprovalRequired bool   `json:"approval_required"`
	Reason           string `json:"reason"`
}

var roleCapabilities = map[string]map[Action]struct{}{
	RoleMaintainer: {
		ActionCaseCreate:           {},
		ActionCaseView:             {},
		ActionRecommendationReview: {},
		ActionEvidenceView:         {},
	},
	RoleProductionController: {
		ActionCaseCreate:           {},
		ActionCaseView:             {},
		ActionCaseTransition:       {},
		ActionBlockerUpdate:        {},
		ActionRecommendationReview: {},
		ActionApprovalRequest:      {},
		ActionEvidenceView:         {},
	},
	RoleSupplyAnalyst: {
		ActionCaseView:          {},
		ActionPartConstraintAdd: {},
		ActionEvidenceView:      {},
	},
	RoleSustainmentEngineer: {
		ActionCaseView:        {},
		ActionProcedureLink:   {},
		ActionBlockerUpdate:   {},
		ActionApprovalRequest: {},
		ActionEvidenceView:    {},
	},
	RoleSupervisor: {
		ActionCaseView:             {},
		ActionCaseTransition:       {},
		ActionBlockerUpdate:        {},
		ActionProcedureLink:        {},
		ActionPartConstraintAdd:    {},
		ActionRecommendationReview: {},
		ActionApprovalRequest:      {},
		ActionApprovalDecide:       {},
		ActionEvidenceView:         {},
	},
	RolePolicyReviewer: {
		ActionCaseView:             {},
		ActionApprovalDecide:       {},
		ActionEvidenceView:         {},
		ActionRecommendationReview: {},
	},
	RoleAdministrator: {
		ActionCaseCreate:           {},
		ActionCaseView:             {},
		ActionCaseTransition:       {},
		ActionBlockerUpdate:        {},
		ActionProcedureLink:        {},
		ActionPartConstraintAdd:    {},
		ActionRecommendationReview: {},
		ActionApprovalRequest:      {},
		ActionApprovalDecide:       {},
		ActionEvidenceView:         {},
	},
}

// Evaluate returns a policy decision for a requested action. The goal is to
// keep authorization logic explicit, conservative, and inspectable.
func Evaluate(input GuardInput) Decision {
	if input.ActorRole == "" {
		return deny("actor role is required")
	}

	if input.Action == "" {
		return deny("action is required")
	}

	if !roleAllows(input.ActorRole, input.Action) {
		return deny(fmt.Sprintf("role %q is not allowed to perform action %q", input.ActorRole, input.Action))
	}

	switch input.Action {
	case ActionCaseTransition:
		return evaluateCaseTransition(input)
	case ActionProcedureLink:
		return evaluateProcedureLink(input)
	case ActionRecommendationReview:
		return evaluateRecommendationReview(input)
	case ActionApprovalDecide:
		return evaluateApprovalDecision(input)
	default:
		return allow("action allowed without additional approval")
	}
}

func evaluateCaseTransition(input GuardInput) Decision {
	if err := validateStateTransitionInputs(input.CurrentState, input.TargetState); err != nil {
		return deny(err.Error())
	}

	switch input.TargetState {
	case domain.CaseStateClosed:
		if input.ActorRole != RoleSupervisor && input.ActorRole != RoleAdministrator {
			return deny("closing a case requires supervisor or administrator role")
		}
	case domain.CaseStateResolved:
		if input.ActorRole == RoleMaintainer {
			return requireApproval("resolving a case requires accountable review beyond maintainer role")
		}
	case domain.CaseStateActionable:
		if input.RestrictedProcedure || input.RecommendationNeedsApproval || input.OverrideRequested {
			return requireApproval("marking a case actionable requires approval when restricted data, override behavior, or approval-gated recommendations are involved")
		}
	}

	return allow("state transition allowed")
}

func evaluateProcedureLink(input GuardInput) Decision {
	if input.RestrictedProcedure {
		if input.ActorRole == RoleSupervisor || input.ActorRole == RoleAdministrator {
			return requireApproval("restricted procedure linkage requires recorded approval even for elevated roles")
		}

		return deny("restricted procedure linkage requires supervisor-approved access path")
	}

	return allow("procedure linkage allowed")
}

func evaluateRecommendationReview(input GuardInput) Decision {
	if input.OverrideRequested {
		if input.ActorRole == RoleSupervisor || input.ActorRole == RoleAdministrator {
			return requireApproval("recommendation override requires accountable approval record")
		}

		return deny("recommendation override is restricted to supervisor or administrator role")
	}

	if input.RecommendationNeedsApproval {
		return requireApproval("this recommendation class requires approval before it can influence protected workflow actions")
	}

	switch input.RecommendationStatus {
	case domain.RecommendationStatusAccepted,
		domain.RecommendationStatusRejected,
		domain.RecommendationStatusOverridden:
		return deny("recommendation has already reached a terminal reviewed state")
	default:
		return allow("recommendation review allowed")
	}
}

func evaluateApprovalDecision(input GuardInput) Decision {
	if input.ActorRole != RoleSupervisor && input.ActorRole != RolePolicyReviewer && input.ActorRole != RoleAdministrator {
		return deny("only supervisor, policy-reviewer, or administrator roles may decide approvals")
	}

	return allow("approval decision allowed")
}

func validateStateTransitionInputs(currentState, targetState domain.CaseState) error {
	if currentState == "" {
		return fmt.Errorf("current state is required")
	}
	if targetState == "" {
		return fmt.Errorf("target state is required")
	}

	switch currentState {
	case domain.CaseStateNew,
		domain.CaseStateTriage,
		domain.CaseStateAwaitingData,
		domain.CaseStateAwaitingParts,
		domain.CaseStateAwaitingApproval,
		domain.CaseStateActionable,
		domain.CaseStateDeferred,
		domain.CaseStateResolved,
		domain.CaseStateClosed:
	default:
		return fmt.Errorf("unknown current state %q", currentState)
	}

	switch targetState {
	case domain.CaseStateNew,
		domain.CaseStateTriage,
		domain.CaseStateAwaitingData,
		domain.CaseStateAwaitingParts,
		domain.CaseStateAwaitingApproval,
		domain.CaseStateActionable,
		domain.CaseStateDeferred,
		domain.CaseStateResolved,
		domain.CaseStateClosed:
	default:
		return fmt.Errorf("unknown target state %q", targetState)
	}

	return nil
}

func roleAllows(role string, action Action) bool {
	capabilities, ok := roleCapabilities[role]
	if !ok {
		return false
	}

	_, ok = capabilities[action]
	return ok
}

func allow(reason string) Decision {
	return Decision{
		Allowed:          true,
		ApprovalRequired: false,
		Reason:           reason,
	}
}

func requireApproval(reason string) Decision {
	return Decision{
		Allowed:          true,
		ApprovalRequired: true,
		Reason:           reason,
	}
}

func deny(reason string) Decision {
	return Decision{
		Allowed:          false,
		ApprovalRequired: false,
		Reason:           reason,
	}
}
