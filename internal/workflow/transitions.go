package workflow

import (
	"fmt"

	"github.com/BryceWDesign/IX-Sustainment-OS/internal/domain"
)

var allowedTransitions = map[domain.CaseState]map[domain.CaseState]struct{}{
	domain.CaseStateNew: {
		domain.CaseStateTriage: {},
	},
	domain.CaseStateTriage: {
		domain.CaseStateAwaitingData:     {},
		domain.CaseStateAwaitingParts:    {},
		domain.CaseStateAwaitingApproval: {},
		domain.CaseStateActionable:       {},
		domain.CaseStateDeferred:         {},
	},
	domain.CaseStateAwaitingData: {
		domain.CaseStateTriage:     {},
		domain.CaseStateActionable: {},
		domain.CaseStateDeferred:   {},
	},
	domain.CaseStateAwaitingParts: {
		domain.CaseStateTriage:     {},
		domain.CaseStateActionable: {},
		domain.CaseStateDeferred:   {},
	},
	domain.CaseStateAwaitingApproval: {
		domain.CaseStateActionable: {},
		domain.CaseStateTriage:     {},
		domain.CaseStateDeferred:   {},
	},
	domain.CaseStateActionable: {
		domain.CaseStateResolved: {},
		domain.CaseStateTriage:   {},
		domain.CaseStateDeferred: {},
	},
	domain.CaseStateDeferred: {
		domain.CaseStateTriage: {},
		domain.CaseStateClosed: {},
	},
	domain.CaseStateResolved: {
		domain.CaseStateClosed: {},
		domain.CaseStateTriage: {},
	},
	domain.CaseStateClosed: {
		domain.CaseStateTriage: {},
	},
}

// AllowedNextStates returns the valid next states from the provided state
// in a stable, operator-readable order.
func AllowedNextStates(from domain.CaseState) []domain.CaseState {
	switch from {
	case domain.CaseStateNew:
		return []domain.CaseState{
			domain.CaseStateTriage,
		}
	case domain.CaseStateTriage:
		return []domain.CaseState{
			domain.CaseStateAwaitingData,
			domain.CaseStateAwaitingParts,
			domain.CaseStateAwaitingApproval,
			domain.CaseStateActionable,
			domain.CaseStateDeferred,
		}
	case domain.CaseStateAwaitingData:
		return []domain.CaseState{
			domain.CaseStateTriage,
			domain.CaseStateActionable,
			domain.CaseStateDeferred,
		}
	case domain.CaseStateAwaitingParts:
		return []domain.CaseState{
			domain.CaseStateTriage,
			domain.CaseStateActionable,
			domain.CaseStateDeferred,
		}
	case domain.CaseStateAwaitingApproval:
		return []domain.CaseState{
			domain.CaseStateActionable,
			domain.CaseStateTriage,
			domain.CaseStateDeferred,
		}
	case domain.CaseStateActionable:
		return []domain.CaseState{
			domain.CaseStateResolved,
			domain.CaseStateTriage,
			domain.CaseStateDeferred,
		}
	case domain.CaseStateDeferred:
		return []domain.CaseState{
			domain.CaseStateTriage,
			domain.CaseStateClosed,
		}
	case domain.CaseStateResolved:
		return []domain.CaseState{
			domain.CaseStateClosed,
			domain.CaseStateTriage,
		}
	case domain.CaseStateClosed:
		return []domain.CaseState{
			domain.CaseStateTriage,
		}
	default:
		return nil
	}
}

// CanTransition reports whether a state transition is valid.
func CanTransition(from, to domain.CaseState) bool {
	next, ok := allowedTransitions[from]
	if !ok {
		return false
	}

	_, ok = next[to]
	return ok
}

// ValidateTransition returns an error when a transition is invalid.
func ValidateTransition(from, to domain.CaseState) error {
	if from == "" {
		return fmt.Errorf("source state is required")
	}
	if to == "" {
		return fmt.Errorf("target state is required")
	}
	if !CanTransition(from, to) {
		return fmt.Errorf("invalid state transition: %s -> %s", from, to)
	}

	return nil
}

// PrimaryBlocker returns the explicitly marked primary blocker when present.
// If none is flagged as primary, it returns the first blocker in the list.
func PrimaryBlocker(blockers []domain.Blocker) (domain.BlockerCategory, bool) {
	for _, blocker := range blockers {
		if blocker.IsPrimary {
			return blocker.Category, true
		}
	}

	if len(blockers) == 0 {
		return "", false
	}

	return blockers[0].Category, true
}

// HasBlockingCategory reports whether the case currently includes a blocker of
// the specified category.
func HasBlockingCategory(blockers []domain.Blocker, category domain.BlockerCategory) bool {
	for _, blocker := range blockers {
		if blocker.Category == category {
			return true
		}
	}

	return false
}

// DeriveSuggestedState produces a conservative workflow suggestion based on the
// current blockers and approval boundary. It does not replace human judgment;
// it provides a safe default routing hint for the application layer.
func DeriveSuggestedState(blockers []domain.Blocker, approvalRequired bool) domain.CaseState {
	if approvalRequired || HasBlockingCategory(blockers, domain.BlockerCategoryApproval) {
		return domain.CaseStateAwaitingApproval
	}

	primary, ok := PrimaryBlocker(blockers)
	if !ok {
		return domain.CaseStateActionable
	}

	switch primary {
	case domain.BlockerCategoryData,
		domain.BlockerCategoryProcedure,
		domain.BlockerCategoryEntitlement:
		return domain.CaseStateAwaitingData
	case domain.BlockerCategoryParts,
		domain.BlockerCategoryTooling:
		return domain.CaseStateAwaitingParts
	case domain.BlockerCategoryApproval:
		return domain.CaseStateAwaitingApproval
	case domain.BlockerCategoryCapacity,
		domain.BlockerCategoryPolicy:
		return domain.CaseStateDeferred
	default:
		return domain.CaseStateTriage
	}
}
