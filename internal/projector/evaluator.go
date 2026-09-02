package projector

import (
	"errors"
	"fmt"
	"sort"
)

const (
	unknownIncompleteSeed = "TYPED_REPAIR_SEED_INCOMPLETE"
	unknownStaleEvidence  = "STALE_GRAPH_EVIDENCE"
	unknownTerminal       = "TERMINAL_STATE_NOT_EXPLICIT_FIXED_POINT"
	unknownAmbiguousSeed  = "AMBIGUOUS_TYPED_REPAIR_SEED"
	refutedGraph          = "IMMUTABLE_GRAPH_REQUIRED"
	refutedEvidence      = "KNOWN_EVIDENCE_CONTRADICTION"
	refutedAuthority      = "PROPOSAL_AUTHORITY_BOUNDARY_VIOLATED"
	refutedCounterexample = "COUNTEREXAMPLE_IDENTITY_CONTRADICTION"
	refutedOperational    = "ACCIDENTAL_EXECUTION"
)

type seedResult int

const (
	seedValid seedResult = iota
	seedUnknown
	seedRefuted
)

func Evaluate(input Input, inputDigest string, ir SemanticIR) (Dossier, []ProposalEvent, error) {
	if input.Schema != ProtocolSchema+"/input/v1" || input.CaseID == "" {
		return Dossier{}, nil, errors.New("INVALID_INPUT_HEADER")
	}

	nodes := map[string]SemanticNode{}
	for _, node := range input.Graph.Nodes {
		if node.ID == "" || nodes[node.ID].ID != "" || !validDigest(node.Digest) {
			return Dossier{}, nil, fmt.Errorf("INVALID_GRAPH_NODE_%s", node.ID)
		}
		nodes[node.ID] = node
	}
	refutations := make([]RefutedIncident, 0)
	frontier := make([]FrontierItem, 0)
	validSeeds := make([]struct {
		evidence EvidenceRecord
		seed     RepairSeed
	}, 0)
	seenEvidence := map[string]bool{}
	unknownAdded := false
	addUnknown := func(kind string, record UnknownRecord, targetIDs []string) {
		if unknownAdded {
			return
		}
		unknownAdded = true
		frontier = append(frontier, FrontierItem{
			ID: "frontier." + input.CaseID + "." + kind,
			Kind: kind,
			TargetSemanticIDs: append([]string(nil), targetIDs...),
			Unknown: record,
		})
	}
	addRefuted := func(id, kind, evidenceID, operation, reason string, targetIDs []string) {
		refutations = append(refutations, RefutedIncident{
			ID: id, Kind: kind, EvidenceID: evidenceID, TargetSemanticIDs: append([]string(nil), targetIDs...), Operation: operation, Reason: reason,
		})
	}

	if input.Graph.Schema != "gooo/semantic-graph/v1" || !input.Graph.Immutable || !validDigest(input.Graph.Digest) {
		addRefuted("incident.graph.mutable", "SEMANTIC_REFUTED", "", "BIND_IMMUTABLE_GRAPH", refutedGraph, nil)
	}
	if input.Graph.Identity.Repository == "" || input.Graph.Identity.ReleaseID == "" || input.Graph.Identity.Tag == "" || !validCommit(input.Graph.Identity.CommitSHA) || input.Graph.Identity.Asset == "" || !validDigest(input.Graph.Identity.AssetDigest) {
		addRefuted("incident.graph.identity", "SEMANTIC_REFUTED", "", "BIND_RELEASE_IDENTITY", "GRAPH_RELEASE_IDENTITY_INVALID", nil)
	}
	if input.Evidence.Schema != "gooo/conformance-evidence/v1" || !input.Evidence.Immutable || !validDigest(input.Evidence.Digest) {
		addRefuted("incident.evidence.immutable", "SEMANTIC_REFUTED", "", "BIND_IMMUTABLE_EVIDENCE", refutedEvidence, nil)
	}
	if input.Evidence.GraphDigest == "" || input.Evidence.GraphDigest != input.Graph.Digest {
		if input.Evidence.Immutable && validDigest(input.Evidence.Digest) {
			addUnknown(unknownStaleEvidence, UnknownRecord{Stage: "foundation", Step: "evidence_binding", Reason: unknownStaleEvidence, UnknownClass: "STALE", NextOperation: "REFRESH_EVIDENCE_FOR_GRAPH_DIGEST", BlockedBy: "semantic_graph.digest"}, nil)
		}
	}

	if input.Unknown != nil {
		if err := validateUnknown(*input.Unknown); err != nil {
			addRefuted("incident.unknown.malformed", "SEMANTIC_REFUTED", "", "PRESERVE_UNKNOWN_FRONTIER", "UNKNOWN_RECORD_NOT_EXACTLY_SIX_FIELDS", nil)
		} else {
			addUnknown("explicit", *input.Unknown, nil)
		}
	}

	for _, record := range input.Evidence.Records {
		if seenEvidence[record.ID] || record.ID == "" {
			addRefuted("incident.evidence.duplicate", "SEMANTIC_REFUTED", record.ID, "BIND_EVIDENCE_RECORD", refutedEvidence, record.TargetSemanticIDs)
			continue
		}
		seenEvidence[record.ID] = true
		if record.Status != "FAILED" && record.Status != "REFUTED" && record.Status != "CLOSED" {
			addRefuted("incident.evidence.status."+record.ID, "SEMANTIC_REFUTED", record.ID, "CLASSIFY_EVIDENCE", refutedEvidence, record.TargetSemanticIDs)
			continue
		}
		if !record.Immutable {
			addRefuted("incident.evidence.mutable."+record.ID, "SEMANTIC_REFUTED", record.ID, "BIND_EVIDENCE_RECORD", refutedEvidence, record.TargetSemanticIDs)
		}
		if record.GraphDigest == "" || record.GraphDigest != input.Graph.Digest {
			if record.Unknown != nil {
				if err := validateUnknown(*record.Unknown); err != nil {
					addRefuted("incident.unknown.malformed."+record.ID, "SEMANTIC_REFUTED", record.ID, "PRESERVE_UNKNOWN_FRONTIER", "UNKNOWN_RECORD_NOT_EXACTLY_SIX_FIELDS", record.TargetSemanticIDs)
				} else {
					addUnknown("record-"+record.ID, *record.Unknown, record.TargetSemanticIDs)
				}
			} else {
				addUnknown(unknownStaleEvidence, UnknownRecord{Stage: "foundation", Step: "record_binding", Reason: unknownStaleEvidence, UnknownClass: "STALE", NextOperation: "REFRESH_EVIDENCE_FOR_GRAPH_DIGEST", BlockedBy: record.ID}, record.TargetSemanticIDs)
			}
			continue
		}
		if !validProof(record.ProofChoice) || !validIndicator(record.IndicatorClass) {
			addRefuted("incident.evidence.classification."+record.ID, "SEMANTIC_REFUTED", record.ID, "BIND_PROOF_AND_INDICATOR", refutedEvidence, record.TargetSemanticIDs)
		}
		if !allNodesExist(record.TargetSemanticIDs, nodes) {
			addRefuted("incident.evidence.target."+record.ID, "SEMANTIC_REFUTED", record.ID, "BIND_TARGET_SEMANTIC_IDS", refutedEvidence, record.TargetSemanticIDs)
			continue
		}
		if record.Status == "REFUTED" {
			addRefuted("incident.evidence.refuted."+record.ID, "SEMANTIC_REFUTED", record.ID, "PRESERVE_REFUTED_EVIDENCE", refutedEvidence, record.TargetSemanticIDs)
			continue
		}
		if record.Status != "FAILED" {
			continue
		}
		if record.Repair == nil {
			if record.Unknown != nil {
				if err := validateUnknown(*record.Unknown); err != nil {
					addRefuted("incident.unknown.malformed."+record.ID, "SEMANTIC_REFUTED", record.ID, "PRESERVE_UNKNOWN_FRONTIER", "UNKNOWN_RECORD_NOT_EXACTLY_SIX_FIELDS", record.TargetSemanticIDs)
				} else {
					addUnknown("record-"+record.ID, *record.Unknown, record.TargetSemanticIDs)
				}
			} else {
				addUnknown(unknownIncompleteSeed, UnknownRecord{Stage: "coherence", Step: "seed_validation", Reason: unknownIncompleteSeed, UnknownClass: "MISSING", NextOperation: "SUPPLY_TYPED_REPAIR_SEED", BlockedBy: record.ID}, record.TargetSemanticIDs)
			}
			continue
		}
		result := validateRepairSeed(*record.Repair, record, nodes)
		switch result {
		case seedUnknown:
			if record.Unknown != nil {
				if err := validateUnknown(*record.Unknown); err != nil {
					addRefuted("incident.unknown.malformed."+record.ID, "SEMANTIC_REFUTED", record.ID, "PRESERVE_UNKNOWN_FRONTIER", "UNKNOWN_RECORD_NOT_EXACTLY_SIX_FIELDS", record.TargetSemanticIDs)
				} else {
					addUnknown("record-"+record.ID, *record.Unknown, record.TargetSemanticIDs)
				}
			} else {
				addUnknown(unknownIncompleteSeed, UnknownRecord{Stage: "coherence", Step: "seed_validation", Reason: unknownIncompleteSeed, UnknownClass: "INCOMPLETE", NextOperation: "SUPPLY_TYPED_REPAIR_SEED", BlockedBy: record.ID}, record.TargetSemanticIDs)
			}
		case seedRefuted:
			addRefuted("incident.authority.overreach", "SEMANTIC_REFUTED", record.ID, "EMIT_PROPOSAL_ONLY", refutedAuthority, record.TargetSemanticIDs)
		case seedValid:
			if record.CounterexampleID != "" && input.Counterexample == nil {
				addUnknown("counterexample-"+record.ID, UnknownRecord{Stage: "coherence", Step: "counterexample_binding", Reason: "COUNTEREXAMPLE_REPLAY_EVIDENCE_MISSING", UnknownClass: "MISSING", NextOperation: "SUPPLY_COUNTEREXAMPLE", BlockedBy: record.ID}, record.TargetSemanticIDs)
			} else {
				validSeeds = append(validSeeds, struct {
					evidence EvidenceRecord
					seed     RepairSeed
				}{evidence: record, seed: *record.Repair})
			}
		}
	}

	if input.Counterexample != nil {
		validateCounterexample(*input.Counterexample, input, nodes, seenEvidence, addRefuted)
	}
	for _, operational := range input.OperationalEvents {
		if operational.ID == "" {
			addRefuted("incident.operational.missing-id", "OPERATIONAL_REFUTED", "", operational.Kind, refutedOperational, nil)
			continue
		}
		if operational.Executed || operational.Kind == forbiddenApply || operational.Kind == forbiddenCommit || operational.Kind == forbiddenMerge || operational.Kind == forbiddenWrite {
			addRefuted("incident.operational."+operational.ID, "OPERATIONAL_REFUTED", "", operational.Kind, refutedOperational, []string{operational.Target})
		}
	}
	if input.Evidence.TerminalState != "FIXED_POINT" {
		addUnknown(unknownTerminal, UnknownRecord{Stage: "regression", Step: "terminal_state", Reason: unknownTerminal, UnknownClass: "AMBIGUOUS", NextOperation: "RECORD_EXPLICIT_FIXED_POINT", BlockedBy: "evidence.terminal_state"}, nil)
	}
	if len(validSeeds) > 1 {
		addUnknown(unknownAmbiguousSeed, UnknownRecord{Stage: "coherence", Step: "seed_selection", Reason: unknownAmbiguousSeed, UnknownClass: "AMBIGUOUS", NextOperation: "DISCHARGE_SEED_AMBIGUITY", BlockedBy: "evidence.records"}, nil)
	}

	state := StateClosed
	if len(frontier) > 0 {
		state = StateUnknown
	}
	if len(refutations) > 0 {
		state = StateRefuted
	}
	var proposal *RepairProposal
	if state == StateClosed && len(validSeeds) == 1 {
		proposal = buildProposal(input, validSeeds[0].evidence, validSeeds[0].seed)
	}
	proofVector, indicatorVector := exactBucketVectors(ir.Activities)
	improvementState, improvement := pairImprovement(input)
	vector := ExactVector{
		UnresolvedFrontier: len(frontier),
		Incidents: len(refutations),
	}
	if proposal != nil {
		vector = ExactVector{
			TargetSemanticIDs: len(proposal.TargetSemanticIDs), AllowedOperations: len(proposal.AllowedOperations), ForbiddenOperations: len(proposal.ForbiddenOperations),
			Preconditions: len(proposal.Preconditions), ClaimedChangedCells: len(proposal.ClaimedChangedCells), UnchangedBoundary: len(proposal.UnchangedBoundary),
			ExpectedEvidence: len(proposal.ExpectedEvidence), ValidationPlan: len(proposal.ValidationPlan), CapabilityBudget: len(proposal.CapabilityEffectBudget.Capabilities),
			EffectBudget: len(proposal.CapabilityEffectBudget.Effects), UnresolvedFrontier: len(frontier), Incidents: len(refutations),
		}
	}
	dossier := Dossier{
		Schema: ProtocolSchema + "/dossier/v1", CaseID: input.CaseID, InputDigest: inputDigest, GraphDigest: input.Graph.Digest,
		State: state, Decision: string(state), ResolutionPrecedence: []State{StateRefuted, StateUnknown, StateClosed}, Proposal: proposal,
		UnknownFrontier: frontier, RefutedIncidents: refutations, ExactVector: vector, ProofVector: proofVector, IndicatorVector: indicatorVector,
		ImprovementState: improvementState, Improvement: improvement, Authority: fixedAuthority(),
	}
	if state == StateClosed {
		dossier.Decision = "FIXED_POINT"
	}
	events := buildEvents(input, dossier)
	return dossier, events, nil
}

func validateRepairSeed(seed RepairSeed, evidence EvidenceRecord, nodes map[string]SemanticNode) seedResult {
	if len(seed.TargetSemanticIDs) == 0 || !allNodesExist(seed.TargetSemanticIDs, nodes) || !sameStrings(seed.TargetSemanticIDs, evidence.TargetSemanticIDs) {
		return seedRefuted
	}
	if len(seed.AllowedOperations) == 0 || len(seed.Preconditions) == 0 || len(seed.ClaimedChangedCells) == 0 || len(seed.UnchangedBoundary) == 0 || len(seed.ExpectedEvidence) == 0 || len(seed.ValidationPlan) == 0 || len(seed.CapabilityEffectBudget.Capabilities) == 0 || len(seed.CapabilityEffectBudget.Effects) == 0 {
		return seedUnknown
	}
	if len(seed.AllowedOperations) != 1 || seed.AllowedOperations[0].Kind != proposalOperationKind || seed.AllowedOperations[0].Effect != "PROPOSAL_ONLY" || !sameStrings(seed.AllowedOperations[0].TargetSemanticIDs, seed.TargetSemanticIDs) {
		return seedRefuted
	}
	if len(seed.ForbiddenOperations) != 6 {
		return seedUnknown
	}
	seenForbidden := map[string]bool{}
	for _, operation := range seed.ForbiddenOperations {
		if operation.ID == "" || seenForbidden[operation.Kind] || !forbiddenKind(operation.Kind) || operation.Effect != "NEVER_EXECUTE" || len(operation.TargetSemanticIDs) != 0 {
			return seedRefuted
		}
		seenForbidden[operation.Kind] = true
	}
	if len(seenForbidden) != 6 {
		return seedRefuted
	}
	seenIDs := map[string]bool{}
	for _, id := range seed.TargetSemanticIDs {
		if seenIDs[id] {
			return seedRefuted
		}
		seenIDs[id] = true
	}
	for _, precondition := range seed.Preconditions {
		if precondition.ID == "" || precondition.Kind == "" || precondition.RequiredValue == "" || len(precondition.TargetSemanticIDs) == 0 || !allNodesExist(precondition.TargetSemanticIDs, nodes) {
			return seedRefuted
		}
	}
	for _, cell := range seed.ClaimedChangedCells {
		node, ok := nodes[cell.SemanticID]
		if !ok || cell.ContractCellID == "" || cell.ContractCellID != node.ContractCellID || !validDigest(cell.BeforeDigest) || !validDigest(cell.AfterDigest) || cell.BeforeDigest != node.Digest || cell.BeforeDigest == cell.AfterDigest || cell.EvidenceID != evidence.ID {
			return seedRefuted
		}
	}
	for _, boundary := range seed.UnchangedBoundary {
		if boundary.ID == "" || boundary.Kind == "" || !boundary.MustRemainUnchanged || len(boundary.SemanticIDs) == 0 || !allNodesExist(boundary.SemanticIDs, nodes) {
			return seedRefuted
		}
	}
	for _, expected := range seed.ExpectedEvidence {
		if expected.ID == "" || expected.Kind == "" || !validState(State(expected.RequiredStatus)) || len(expected.TargetSemanticIDs) == 0 || !allNodesExist(expected.TargetSemanticIDs, nodes) || len(expected.ReceiptFields) == 0 {
			return seedRefuted
		}
	}
	for _, step := range seed.ValidationPlan {
		if step.ID == "" || step.Kind == "" || len(step.TargetSemanticIDs) == 0 || !allNodesExist(step.TargetSemanticIDs, nodes) || len(step.RequiredEvidence) == 0 || step.Authority == "" || step.Authority == "runtime_apply" {
			return seedRefuted
		}
	}
	if len(seed.CapabilityEffectBudget.Capabilities) != 2 || len(seed.CapabilityEffectBudget.Effects) != 2 {
		return seedUnknown
	}
	for _, budget := range seed.CapabilityEffectBudget.Capabilities {
		if budget.ID == "" || budget.Limit < 0 || budget.MayApply {
			return seedRefuted
		}
	}
	for _, budget := range seed.CapabilityEffectBudget.Effects {
		if budget.ID == "" || budget.Limit < 0 || budget.MayApply {
			return seedRefuted
		}
	}
	return seedValid
}

func validateCounterexample(counterexample Counterexample, input Input, nodes map[string]SemanticNode, evidence map[string]bool, addRefuted func(string, string, string, string, string, []string)) {
	if counterexample.ID == "" || !validDigest(counterexample.GraphDigest) || counterexample.GraphDigest != input.Graph.Digest || counterexample.FailureEvidenceID == "" || !evidence[counterexample.FailureEvidenceID] || len(counterexample.TargetSemanticIDs) == 0 || !allNodesExist(counterexample.TargetSemanticIDs, nodes) || counterexample.ObservedValue == "" || counterexample.ExpectedValue == "" || counterexample.ObservedValue == counterexample.ExpectedValue || !validDigest(counterexample.ObservedDigest) || !validDigest(counterexample.ExpectedDigest) || counterexample.ObservedDigest == counterexample.ExpectedDigest {
		addRefuted("incident.counterexample."+counterexample.ID, "SEMANTIC_REFUTED", counterexample.FailureEvidenceID, "BIND_COUNTEREXAMPLE", refutedCounterexample, counterexample.TargetSemanticIDs)
	}
}

func validateUnknown(value UnknownRecord) error {
	if value.Stage == "" || value.Step == "" || value.Reason == "" || value.UnknownClass == "" || value.NextOperation == "" || value.BlockedBy == "" {
		return errors.New("UNKNOWN_RECORD_INCOMPLETE")
	}
	return nil
}

func buildProposal(input Input, evidence EvidenceRecord, seed RepairSeed) *RepairProposal {
	return &RepairProposal{
		Schema: ProtocolSchema + "/proposal/v1", ProposalID: "proposal." + input.CaseID + "." + evidence.ID, SourceEvidenceID: evidence.ID, SourceGraphDigest: input.Graph.Digest,
		TargetSemanticIDs: append([]string(nil), seed.TargetSemanticIDs...), AllowedOperations: append([]Operation(nil), seed.AllowedOperations...), ForbiddenOperations: append([]Operation(nil), seed.ForbiddenOperations...),
		Preconditions: append([]Precondition(nil), seed.Preconditions...), ClaimedChangedCells: append([]ChangedCell(nil), seed.ClaimedChangedCells...), UnchangedBoundary: append([]BoundaryClaim(nil), seed.UnchangedBoundary...),
		ExpectedEvidence: append([]EvidenceExpectation(nil), seed.ExpectedEvidence...), ValidationPlan: append([]ValidationStep(nil), seed.ValidationPlan...), CapabilityEffectBudget: seed.CapabilityEffectBudget,
		UnresolvedFrontier: []FrontierItem{}, Authority: fixedAuthority(),
	}
}

func pairImprovement(input Input) (State, *ImprovementClaim) {
	if input.Utility == nil || input.Utility.ScenarioID != input.CaseID || input.Utility.GraphDigest != input.Graph.Digest || input.Utility.Toolchain == "" || input.Utility.Runner == "" || input.Utility.MetricID == "" || input.Utility.Before == nil || input.Utility.After == nil {
		return StateUnknown, nil
	}
	if input.Utility.Before.WallMS < 0 || input.Utility.Before.RSSKib < 0 || input.Utility.After.WallMS < 0 || input.Utility.After.RSSKib < 0 {
		return StateRefuted, nil
	}
	return StateClosed, &ImprovementClaim{
		MetricID: input.Utility.MetricID, Before: *input.Utility.Before, After: *input.Utility.After,
		Delta: UtilityVector{WallMS: input.Utility.After.WallMS - input.Utility.Before.WallMS, RSSKib: input.Utility.After.RSSKib - input.Utility.Before.RSSKib},
	}
}

func exactBucketVectors(activities []SourceActivity) (BucketVector, IndicatorVector) {
	proof := BucketVector{}
	indicator := IndicatorVector{}
	for _, activity := range activities {
		switch activity.ProofChoice {
		case "FOUNDATION": proof.FOUNDATION++
		case "COHERENCE": proof.COHERENCE++
		case "REGRESSION": proof.REGRESSION++
		}
		switch activity.IndicatorClass {
		case "DRIVER": indicator.DRIVER++
		case "OUTCOME": indicator.OUTCOME++
		case "GUARDRAIL": indicator.GUARDRAIL++
		}
	}
	return proof, indicator
}

func buildEvents(input Input, dossier Dossier) []ProposalEvent {
	events := make([]ProposalEvent, 0, len(dossier.RefutedIncidents)+len(dossier.UnknownFrontier)+1)
	for index, incident := range dossier.RefutedIncidents {
		events = append(events, ProposalEvent{Schema: ProtocolSchema + "/event/v1", Ordinal: index + 1, EventID: incident.ID, EventType: incident.Kind, CaseID: input.CaseID, EvidenceID: incident.EvidenceID, TargetIDs: append([]string(nil), incident.TargetSemanticIDs...), InputDigest: inputDigestForEvent(input), Reason: incident.Reason})
	}
	for _, item := range dossier.UnknownFrontier {
		events = append(events, ProposalEvent{Schema: ProtocolSchema + "/event/v1", Ordinal: len(events) + 1, EventID: item.ID, EventType: "UNKNOWN_FRONTIER", CaseID: input.CaseID, TargetIDs: append([]string(nil), item.TargetSemanticIDs...), InputDigest: inputDigestForEvent(input), Reason: item.Unknown.Reason})
	}
	if dossier.Proposal != nil {
		events = append(events, ProposalEvent{Schema: ProtocolSchema + "/event/v1", Ordinal: len(events) + 1, EventID: dossier.Proposal.ProposalID, EventType: "PROPOSAL_EMITTED", CaseID: input.CaseID, EvidenceID: dossier.Proposal.SourceEvidenceID, TargetIDs: append([]string(nil), dossier.Proposal.TargetSemanticIDs...), InputDigest: inputDigestForEvent(input), Reason: "TYPED_EVIDENCE_BOUND"})
	} else if len(events) == 0 {
		events = append(events, ProposalEvent{Schema: ProtocolSchema + "/event/v1", Ordinal: 1, EventID: "event." + input.CaseID + ".fixed-point", EventType: "NO_PROPOSAL_FIXED_POINT", CaseID: input.CaseID, InputDigest: inputDigestForEvent(input), Reason: "NO_TYPED_REPAIR_REQUIRED"})
	}
	for index := range events {
		events[index].EventDigest, _ = DigestJSON(struct {
		Schema string `json:"schema"`; Ordinal int `json:"ordinal"`; EventID string `json:"event_id"`; EventType string `json:"event_type"`; CaseID string `json:"case_id"`; EvidenceID string `json:"evidence_id"`; TargetIDs []string `json:"target_semantic_ids"`; InputDigest string `json:"input_digest"`; Reason string `json:"reason"`
	}{events[index].Schema, events[index].Ordinal, events[index].EventID, events[index].EventType, events[index].CaseID, events[index].EvidenceID, events[index].TargetIDs, events[index].InputDigest, events[index].Reason})
	}
	return events
}

func inputDigestForEvent(input Input) string {
	digest, _ := DigestJSON(input)
	return digest
}

func allNodesExist(ids []string, nodes map[string]SemanticNode) bool {
	if len(ids) == 0 {
		return false
	}
	for _, id := range ids {
		if id == "" || nodes[id].ID == "" {
			return false
		}
	}
	return true
}

func ResolveState(states ...State) State {
	for _, state := range []State{StateRefuted, StateUnknown, StateClosed} {
		for _, observed := range states {
			if observed == state {
				return state
			}
		}
	}
	return StateClosed
}

func SortStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
