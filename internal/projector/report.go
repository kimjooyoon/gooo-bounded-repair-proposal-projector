package projector

import (
	"encoding/json"
	"fmt"
	"strings"
)

func RenderHumanReport(dossier Dossier, events []ProposalEvent, replay ReplayReceipt) string {
	var builder strings.Builder
	fmt.Fprintln(&builder, "# Gooo bounded repair proposal dossier")
	fmt.Fprintln(&builder)
	fmt.Fprintf(&builder, "case_id: `%s`  \nstate: `%s`  \ndecision: `%s`  \ngraph_digest: `%s`\n", dossier.CaseID, dossier.State, dossier.Decision, dossier.GraphDigest)
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "The dossier is a projection only. It never applies code or mutates the input repository.")
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "## Exact typed proposal fields")
	fmt.Fprintln(&builder)
	if dossier.Proposal == nil {
		fmt.Fprintln(&builder, "proposal: `null`")
	} else {
		fmt.Fprintf(&builder, "proposal_id: `%s`  \nsource_evidence_id: `%s`  \ntarget_semantic_ids: `%s`\n", dossier.Proposal.ProposalID, dossier.Proposal.SourceEvidenceID, strings.Join(dossier.Proposal.TargetSemanticIDs, ","))
		writeJSONField(&builder, "allowed_operations", dossier.Proposal.AllowedOperations)
		writeJSONField(&builder, "forbidden_operations", dossier.Proposal.ForbiddenOperations)
		writeJSONField(&builder, "preconditions", dossier.Proposal.Preconditions)
		writeJSONField(&builder, "claimed_changed_cells", dossier.Proposal.ClaimedChangedCells)
		writeJSONField(&builder, "unchanged_boundary", dossier.Proposal.UnchangedBoundary)
		writeJSONField(&builder, "expected_evidence", dossier.Proposal.ExpectedEvidence)
		writeJSONField(&builder, "validation_plan", dossier.Proposal.ValidationPlan)
		writeJSONField(&builder, "capability_effect_budget", dossier.Proposal.CapabilityEffectBudget)
		writeJSONField(&builder, "unresolved_frontier", dossier.Proposal.UnresolvedFrontier)
	}
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "## Exact vectors")
	fmt.Fprintln(&builder)
	writeJSONField(&builder, "exact_vector", dossier.ExactVector)
	writeJSONField(&builder, "proof_vector", dossier.ProofVector)
	writeJSONField(&builder, "indicator_vector", dossier.IndicatorVector)
	fmt.Fprintf(&builder, "improvement_state: `%s`  \nimprovement: `%s`\n", dossier.ImprovementState, jsonOrNull(dossier.Improvement))
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "## Unresolved frontier")
	fmt.Fprintln(&builder)
	if len(dossier.UnknownFrontier) == 0 {
		fmt.Fprintln(&builder, "none")
	} else {
		for _, item := range dossier.UnknownFrontier {
			fmt.Fprintf(&builder, "- `%s` kind=`%s` targets=`%s` stage=`%s` step=`%s` reason=`%s` unknown_class=`%s` next_operation=`%s` blocked_by=`%s`\n", item.ID, item.Kind, strings.Join(item.TargetSemanticIDs, ","), item.Unknown.Stage, item.Unknown.Step, item.Unknown.Reason, item.Unknown.UnknownClass, item.Unknown.NextOperation, item.Unknown.BlockedBy)
		}
	}
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "## Refuted incidents")
	fmt.Fprintln(&builder)
	if len(dossier.RefutedIncidents) == 0 {
		fmt.Fprintln(&builder, "none")
	} else {
		for _, incident := range dossier.RefutedIncidents {
			fmt.Fprintf(&builder, "- `%s` kind=`%s` evidence=`%s` operation=`%s` targets=`%s` reason=`%s`\n", incident.ID, incident.Kind, incident.EvidenceID, incident.Operation, strings.Join(incident.TargetSemanticIDs, ","), incident.Reason)
		}
	}
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "## Authority boundary")
	fmt.Fprintln(&builder)
	writeJSONField(&builder, "authority", dossier.Authority)
	fmt.Fprintln(&builder, "proposal generation is separate from apply, commit, and merge; runtime apply authority is exactly `0`.")
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "## Replay and artifacts")
	fmt.Fprintln(&builder)
	fmt.Fprintf(&builder, "deterministic: `%t`  \nevent_count: `%d`  \nproposal_digest: `%s`  \nevents_digest: `%s`\n", replay.Deterministic, len(events), replay.ProposalDigest, replay.EventsDigest)
	for _, artifact := range replay.OutputArtifacts {
		fmt.Fprintf(&builder, "- `%s` `%s`\n", artifact.Path, artifact.Digest)
	}
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "No aggregate score, percentage, title inference, automatic apply, commit, merge, release mutation, or deletion is emitted.")
	return builder.String()
}

func writeJSONField(builder *strings.Builder, name string, value any) {
	raw, _ := json.Marshal(value)
	fmt.Fprintf(builder, "%s: `%s`  \n", name, string(raw))
}

func jsonOrNull(value any) string {
	if value == nil {
		return "null"
	}
	raw, _ := json.Marshal(value)
	return string(raw)
}
