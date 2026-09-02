package projector

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func RunConformance(root, sourcePath, corpusPath, outputDir string) (ConformanceIndex, error) {
	if err := EnsureCallerOwnedOutput(outputDir, root); err != nil {
		return ConformanceIndex{}, err
	}
	sourceRaw, err := os.ReadFile(sourcePath)
	if err != nil {
		return ConformanceIndex{}, err
	}
	ir, contract, err := CompileSource(sourcePath, sourceRaw)
	if err != nil {
		return ConformanceIndex{}, err
	}
	corpus, err := ReadCorpus(corpusPath)
	if err != nil {
		return ConformanceIndex{}, err
	}
	if corpus.Schema != ProtocolSchema+"/corpus/v1" || corpus.DenominatorID != contract.ID || corpus.Total != 12 || len(corpus.Cases) != 12 {
		return ConformanceIndex{}, errors.New("CORPUS_MUST_CONTAIN_EXACTLY_TWELVE_CASES")
	}
	contractByID := map[string]ContractCase{}
	for _, item := range contract.Cases {
		contractByID[item.ID] = item
	}
	states := map[string]int{"normal": 0, "unknown": 0, "refuted": 0}
	results := make([]ConformanceResult, 0, len(corpus.Cases))
	seen := map[string]bool{}
	for _, item := range corpus.Cases {
		if seen[item.CaseID] || item.Ordinal < 1 || item.Ordinal > 12 || item.Class == "" {
			return ConformanceIndex{}, fmt.Errorf("INVALID_CORPUS_CASE_%s", item.CaseID)
		}
		seen[item.CaseID] = true
		contractCase, ok := contractByID[item.CaseID]
		if !ok || contractCase.ExpectedState != item.ExpectedState || !sameInts(contractCase.Vector, item.ExpectedVector) || contractCase.ProofChoice != item.ProofChoice || contractCase.IndicatorClass != item.IndicatorClass || contractCase.Class != item.Class {
			return ConformanceIndex{}, fmt.Errorf("CORPUS_CONTRACT_BINDING_MISMATCH_%s", item.CaseID)
		}
		states[item.Class]++
		inputPath := item.Input
		if !filepath.IsAbs(inputPath) {
			inputPath = filepath.Join(root, inputPath)
		}
		input, inputRaw, err := ReadInput(inputPath)
		if err != nil {
			return ConformanceIndex{}, fmt.Errorf("%s read input: %w", item.CaseID, err)
		}
		if input.CaseID != item.CaseID {
			return ConformanceIndex{}, fmt.Errorf("%s INPUT_CASE_ID_MISMATCH", item.CaseID)
		}
		first, firstEvents, err := evaluateWithIR(input, inputRaw, ir)
		if err != nil {
			return ConformanceIndex{}, fmt.Errorf("%s first evaluation: %w", item.CaseID, err)
		}
		second, secondEvents, err := evaluateWithIR(input, inputRaw, ir)
		if err != nil {
			return ConformanceIndex{}, fmt.Errorf("%s replay evaluation: %w", item.CaseID, err)
		}
		if !sameEvaluation(first, firstEvents, second, secondEvents) {
			return ConformanceIndex{}, fmt.Errorf("%s DETERMINISTIC_REPLAY_MISMATCH", item.CaseID)
		}
		if first.State != item.ExpectedState || vectorSlice(first.ExactVector) == nil || !sameInts(vectorSlice(first.ExactVector), item.ExpectedVector) {
			return ConformanceIndex{}, fmt.Errorf("%s EXPECTATION_MISMATCH", item.CaseID)
		}
		incidentIDs := make([]string, 0, len(first.RefutedIncidents))
		for _, incident := range first.RefutedIncidents {
			incidentIDs = append(incidentIDs, incident.ID)
		}
		if !sameStrings(incidentIDs, item.ExpectedIncidentIDs) {
			return ConformanceIndex{}, fmt.Errorf("%s INCIDENT_EXPECTATION_MISMATCH", item.CaseID)
		}
		caseDir := filepath.Join(outputDir, item.CaseID)
		if err := WriteEvaluation(caseDir, Evaluation{Dossier: first, Events: firstEvents, SemanticIRRaw: mustJSON(ir), GeneratedRaw: GenerateEvaluator(ir, DigestBytes(mustJSON(ir)), DigestBytes(sourceRaw)), Replay: ReplayReceipt{}}); err != nil {
			return ConformanceIndex{}, fmt.Errorf("%s write: %w", item.CaseID, err)
		}
		results = append(results, ConformanceResult{Ordinal: item.Ordinal, CaseID: item.CaseID, Class: item.Class, State: first.State, ExactVector: first.ExactVector, ProofChoice: item.ProofChoice, IndicatorClass: item.IndicatorClass, IncidentIDs: incidentIDs, OutputFiles: len(fixedOutputFiles)})
	}
	if states["normal"] != 4 || states["unknown"] != 4 || states["refuted"] != 4 {
		return ConformanceIndex{}, errors.New("CORPUS_STATE_VECTOR_MUST_BE_4_4_4")
	}
	index := ConformanceIndex{Schema: ProtocolSchema + "/conformance-index/v1", CorpusID: corpus.CorpusID, DenominatorID: corpus.DenominatorID, Total: corpus.Total, States: states, Results: results}
	raw, err := jsonBytes(index)
	if err != nil {
		return ConformanceIndex{}, err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "conformance-index.json"), raw, 0o644); err != nil {
		return ConformanceIndex{}, err
	}
	return index, nil
}

func evaluateWithIR(input Input, inputRaw []byte, ir SemanticIR) (Dossier, []ProposalEvent, error) {
	return Evaluate(input, DigestBytes(inputRaw), ir)
}

func sameEvaluation(left Dossier, leftEvents []ProposalEvent, right Dossier, rightEvents []ProposalEvent) bool {
	leftRaw, _ := json.Marshal(struct {
		Dossier Dossier          `json:"dossier"`
		Events  []ProposalEvent  `json:"events"`
	}{left, leftEvents})
	rightRaw, _ := json.Marshal(struct {
		Dossier Dossier          `json:"dossier"`
		Events  []ProposalEvent  `json:"events"`
	}{right, rightEvents})
	return bytes.Equal(leftRaw, rightRaw)
}

func sameInts(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func vectorSlice(vector ExactVector) []int {
	return []int{vector.TargetSemanticIDs, vector.AllowedOperations, vector.ForbiddenOperations, vector.Preconditions, vector.ClaimedChangedCells, vector.UnchangedBoundary, vector.ExpectedEvidence, vector.ValidationPlan, vector.CapabilityBudget, vector.EffectBudget, vector.UnresolvedFrontier, vector.Incidents}
}

func mustJSON(value any) []byte {
	raw, _ := jsonBytes(value)
	return raw
}
