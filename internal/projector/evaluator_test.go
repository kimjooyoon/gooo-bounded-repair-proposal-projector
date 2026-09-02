package projector

import "testing"

func TestResolveStateUsesFixedPrecedence(t *testing.T) {
	if got := ResolveState(StateClosed, StateUnknown, StateRefuted); got != StateRefuted {
		t.Fatalf("got %s, want REFUTED", got)
	}
	if got := ResolveState(StateClosed, StateUnknown); got != StateUnknown {
		t.Fatalf("got %s, want UNKNOWN", got)
	}
}

func TestUnknownRecordRequiresExactlySixNonEmptyCoordinates(t *testing.T) {
	valid := UnknownRecord{Stage: "foundation", Step: "binding", Reason: "STALE", UnknownClass: "STALE", NextOperation: "REFRESH", BlockedBy: "graph.digest"}
	if err := validateUnknown(valid); err != nil {
		t.Fatalf("valid record rejected: %v", err)
	}
	valid.BlockedBy = ""
	if err := validateUnknown(valid); err == nil {
		t.Fatal("incomplete record accepted")
	}
}

func TestAuthoritySeparatesMutationRights(t *testing.T) {
	authority := fixedAuthority()
	if authority.ProposalGenerationAuthority != 1 || !authorityZero(authority) {
		t.Fatalf("unexpected authority boundary: %+v", authority)
	}
}

func TestExactVectorHasNoAggregateField(t *testing.T) {
	vector := vectorSlice(ExactVector{TargetSemanticIDs: 1, AllowedOperations: 1, ForbiddenOperations: 6, Preconditions: 2, ClaimedChangedCells: 1, UnchangedBoundary: 2, ExpectedEvidence: 2, ValidationPlan: 3, CapabilityBudget: 2, EffectBudget: 2})
	if len(vector) != 12 || vector[2] != 6 || vector[11] != 0 {
		t.Fatalf("unexpected exact vector: %v", vector)
	}
}
