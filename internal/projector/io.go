package projector

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var fixedOutputFiles = []string{
	"proposal.json",
	"proposal-events.ndjson",
	"semantic-ir.json",
	"generated-evaluator.go",
	"replay-receipt.json",
	"human-report.md",
}

func FixedOutputFiles() []string {
	return append([]string(nil), fixedOutputFiles...)
}

func ReadInput(path string) (Input, []byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Input{}, nil, err
	}
	var input Input
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return Input{}, nil, err
	}
	return input, raw, nil
}

func ReadCorpus(path string) (Corpus, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Corpus{}, err
	}
	var corpus Corpus
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&corpus); err != nil {
		return Corpus{}, err
	}
	return corpus, nil
}

func EnsureCallerOwnedOutput(outputDir, sourceRoot string) error {
	if !filepath.IsAbs(outputDir) {
		return errors.New("OUTPUT_MUST_BE_ABSOLUTE")
	}
	absOutput, err := filepath.Abs(outputDir)
	if err != nil {
		return err
	}
	absRoot, err := filepath.Abs(sourceRoot)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(absRoot, absOutput)
	if err != nil {
		return err
	}
	if relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator))) {
		return errors.New("OUTPUT_MUST_BE_OUTSIDE_INPUT_REPOSITORY")
	}
	if info, statErr := os.Stat(absOutput); statErr == nil {
		if !info.IsDir() {
			return errors.New("OUTPUT_PATH_NOT_DIRECTORY")
		}
		entries, readErr := os.ReadDir(absOutput)
		if readErr != nil {
			return readErr
		}
		if len(entries) != 0 {
			return errors.New("OUTPUT_DIRECTORY_MUST_BE_EMPTY")
		}
	} else if !os.IsNotExist(statErr) {
		return statErr
	}
	return os.MkdirAll(absOutput, 0o755)
}

func WriteEvaluation(outputDir string, evaluation Evaluation) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	proposalRaw, err := jsonBytes(evaluation.Dossier)
	if err != nil {
		return err
	}
	eventsRaw, err := eventsBytes(evaluation.Events)
	if err != nil {
		return err
	}
	replay := evaluation.Replay
	replay.Schema = ProtocolSchema + "/replay/v1"
	replay.CaseID = evaluation.Dossier.CaseID
	replay.InputDigest = evaluation.Dossier.InputDigest
	replay.ProposalDigest = DigestBytes(proposalRaw)
	replay.EventsDigest = DigestBytes(eventsRaw)
	replay.RepositoryWrites = evaluation.Dossier.Authority.RepositoryWriteAuthority
	replay.ApplyExecutions = evaluation.Dossier.Authority.ApplyAuthority
	replay.CommitExecutions = evaluation.Dossier.Authority.CommitAuthority
	replay.MergeExecutions = evaluation.Dossier.Authority.MergeAuthority
	replay.Deterministic = true
	files := map[string][]byte{
		"proposal.json": proposalRaw,
		"proposal-events.ndjson": eventsRaw,
		"semantic-ir.json": evaluation.SemanticIRRaw,
		"generated-evaluator.go": evaluation.GeneratedRaw,
		"human-report.md": []byte(RenderHumanReport(evaluation.Dossier, evaluation.Events, replay)),
	}
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	replay.OutputArtifacts = make([]ArtifactBinding, 0, len(paths))
	for _, path := range paths {
		if err := os.WriteFile(filepath.Join(outputDir, path), files[path], 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
		replay.OutputArtifacts = append(replay.OutputArtifacts, ArtifactBinding{Path: path, Digest: DigestBytes(files[path])})
	}
	replayRaw, err := jsonBytes(replay)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "replay-receipt.json"), replayRaw, 0o644); err != nil {
		return fmt.Errorf("write replay-receipt.json: %w", err)
	}
	return nil
}

func jsonBytes(value any) ([]byte, error) {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(raw, '\n'), nil
}

func eventsBytes(events []ProposalEvent) ([]byte, error) {
	var builder strings.Builder
	for _, event := range events {
		raw, err := json.Marshal(event)
		if err != nil {
			return nil, err
		}
		builder.Write(raw)
		builder.WriteByte('\n')
	}
	return []byte(builder.String()), nil
}
