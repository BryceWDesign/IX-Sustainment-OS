package validation_test

import (
	"testing"

	"github.com/BryceWDesign/IX-Sustainment-OS/internal/domain"
	"github.com/BryceWDesign/IX-Sustainment-OS/internal/policy"
	"github.com/BryceWDesign/IX-Sustainment-OS/internal/workflow"
)

func TestValidateTransition_AllowsExpectedPath(t *testing.T) {
	t.Parallel()

	err := workflow.ValidateTransition(domain.CaseStateTriage, domain.CaseStateAwaitingApproval)
	if err != nil {
		t.Fatalf("expected valid transition, got error: %v", err)
	}
}

func TestValidateTransition_RejectsInvalidPath(t *testing.T) {
	t.Parallel()

	err := workflow.ValidateTransition(domain.CaseStateAwaitingParts, domain.CaseStateResolved)
	if err == nil {
		t.Fatal("expected invalid transition to fail, but it passed")
	}
}

func TestDeriveSuggestedState_PrefersApprovalWhenApprovalRequired(t *testing.T) {
	t.Parallel()

	state := workflow.DeriveSuggestedState([]domain.Blocker{
		{
			Category:  domain.BlockerCategoryParts,
			Summary:   "Part unavailable",
			IsPrimary: true,
		},
	}, true)

	if state != domain.CaseStateAwaitingApproval {
		t.Fatalf("expected awaiting-approval, got %s", state)
	}
}

func TestDeriveSuggestedState_MapsPrimaryPartsBlocker(t *testing.T) {
	t.Parallel()

	state := workflow.DeriveSuggestedState([]domain.Blocker{
		{
			Category:  domain.BlockerCategoryParts,
			Summary:   "Seal assembly unavailable",
			IsPrimary: true,
		},
	}, false)

	if state != domain.CaseStateAwaitingParts {
		t.Fatalf("expected awaiting-parts, got %s", state)
	}
}

func TestDeriveSuggestedState_NoBlockersMeansActionable(t *testing.T) {
	t.Parallel()

	state := workflow.DeriveSuggestedState(nil, false)
	if state != domain.CaseStateActionable {
		t.Fatalf("expected actionable, got %s", state)
	}
}

func TestPolicyEvaluate_AllowsProductionControllerTransitionWithoutApproval(t *testing.T) {
	t.Parallel()

	decision := policy.Evaluate(policy.GuardInput{
		ActorRole:    policy.RoleProductionController,
		Action:       policy.ActionCaseTransition,
		CurrentState: domain.CaseStateTriage,
		TargetState:  domain.CaseStateAwaitingParts,
	})

	if !decision.Allowed {
		t.Fatalf("expected transition to be allowed, got denied: %s", decision.Reason)
	}

	if decision.ApprovalRequired {
		t.Fatalf("expected no approval requirement, got: %s", decision.Reason)
	}
}

func TestPolicyEvaluate_RequiresApprovalForActionableWhenRestrictedProcedurePresent(t *testing.T) {
	t.Parallel()

	decision := policy.Evaluate(policy.GuardInput{
		ActorRole:           policy.RoleProductionController,
		Action:              policy.ActionCaseTransition,
		CurrentState:        domain.CaseStateAwaitingApproval,
		TargetState:         domain.CaseStateActionable,
		RestrictedProcedure: true,
	})

	if !decision.Allowed {
		t.Fatalf("expected action to remain allowed with approval requirement, got denied: %s", decision.Reason)
	}

	if !decision.ApprovalRequired {
		t.Fatalf("expected approval requirement, got none: %s", decision.Reason)
	}
}

func TestPolicyEvaluate_DeniesMaintainerCaseClosure(t *testing.T) {
	t.Parallel()

	decision := policy.Evaluate(policy.GuardInput{
		ActorRole:    policy.RoleMaintainer,
		Action:       policy.ActionCaseTransition,
		CurrentState: domain.CaseStateResolved,
		TargetState:  domain.CaseStateClosed,
	})

	if decision.Allowed {
		t.Fatalf("expected closure to be denied for maintainer role")
	}
}

func TestPolicyEvaluate_DeniesRestrictedProcedureLinkForNonSupervisor(t *testing.T) {
	t.Parallel()

	decision := policy.Evaluate(policy.GuardInput{
		ActorRole:           policy.RoleSustainmentEngineer,
		Action:              policy.ActionProcedureLink,
		RestrictedProcedure: true,
	})

	if decision.Allowed {
		t.Fatalf("expected restricted procedure link to be denied for sustainment engineer")
	}
}

func TestPolicyEvaluate_RequiresApprovalForSupervisorRecommendationOverride(t *testing.T) {
	t.Parallel()

	decision := policy.Evaluate(policy.GuardInput{
		ActorRole:            policy.RoleSupervisor,
		Action:               policy.ActionRecommendationReview,
		RecommendationStatus: domain.RecommendationStatusPendingReview,
		OverrideRequested:    true,
	})

	if !decision.Allowed {
		t.Fatalf("expected supervisor override path to be allowed with approval requirement, got denied: %s", decision.Reason)
	}

	if !decision.ApprovalRequired {
		t.Fatalf("expected approval requirement for override path, got none")
	}
}
