package projector

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func CompileSource(sourcePath string, source []byte) (SemanticIR, Contract, error) {
	lines := strings.Split(string(source), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != `@gooo schema="gooo/bounded-repair-proposal-projector/v1"` {
		return SemanticIR{}, Contract{}, errors.New("GOOO_SOURCE_SCHEMA_HEADER_MISSING")
	}
	contract := Contract{
		Schema: SourceSchema,
		ResolutionPrecedence: []State{StateRefuted, StateUnknown, StateClosed},
	}
	for lineNumber, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		switch {
		case strings.HasPrefix(line, "contract "):
			attrs, err := parseAttributes(strings.TrimPrefix(line, "contract "))
			if err != nil {
				return SemanticIR{}, Contract{}, fmt.Errorf("source line %d: %w", lineNumber+1, err)
			}
			contract.ID = attrs["id"]
			contract.Total = atoi(attrs["total"])
			precedence := strings.Split(attrs["resolution_precedence"], ">")
			contract.ResolutionPrecedence = make([]State, 0, len(precedence))
			for _, value := range precedence {
				contract.ResolutionPrecedence = append(contract.ResolutionPrecedence, State(value))
			}
			contract.ClosureState = attrs["closure_state"]
			contract.UnknownFields = splitComma(attrs["unknown_fields"])
			contract.NoAggregateScore = attrs["no_aggregate_score"] == "true"
			contract.NoNaturalLanguageInference = attrs["no_natural_language_inference"] == "true"
			contract.RuntimeApplyAuthority = atoi(attrs["runtime_apply_authority"])
			contract.AuthoritySeparation = splitComma(attrs["authority_separation"])
		case strings.HasPrefix(line, "rule "):
			attrs, err := parseAttributes(strings.TrimPrefix(line, "rule "))
			if err != nil {
				return SemanticIR{}, Contract{}, fmt.Errorf("source line %d: %w", lineNumber+1, err)
			}
			if attrs["name"] == "improvement" && attrs["value"] != "exact_matched_before_after_or_null_unknown" {
				return SemanticIR{}, Contract{}, errors.New("IMPROVEMENT_RULE_MISMATCH")
			}
			if attrs["name"] == "proposal" && attrs["value"] != "typed_evidence_only_never_apply" {
				return SemanticIR{}, Contract{}, errors.New("PROPOSAL_RULE_MISMATCH")
			}
		case strings.HasPrefix(line, "activity "):
			attrs, err := parseAttributes(strings.TrimPrefix(line, "activity "))
			if err != nil {
				return SemanticIR{}, Contract{}, fmt.Errorf("source line %d: %w", lineNumber+1, err)
			}
			contract.Activities = append(contract.Activities, SourceActivity{
				ID: attrs["id"], Activity: attrs["activity"], ClaimID: attrs["claim_id"], OperationID: attrs["operation_id"],
				Stage: attrs["stage"], Step: attrs["step"], ProofChoice: attrs["proof_choice"], IndicatorClass: attrs["indicator"], SourceLine: lineNumber + 1,
			})
		case strings.HasPrefix(line, "case "):
			attrs, err := parseAttributes(strings.TrimPrefix(line, "case "))
			if err != nil {
				return SemanticIR{}, Contract{}, fmt.Errorf("source line %d: %w", lineNumber+1, err)
			}
			contract.Cases = append(contract.Cases, ContractCase{
				Ordinal: atoi(attrs["ordinal"]), ID: attrs["id"], Class: attrs["class"], ExpectedState: State(attrs["expected_state"]),
				ProofChoice: attrs["proof_choice"], IndicatorClass: attrs["indicator"], Vector: parseVector(attrs["vector"]),
			})
		default:
			return SemanticIR{}, Contract{}, fmt.Errorf("source line %d: UNKNOWN_GOOO_DIRECTIVE", lineNumber+1)
		}
	}
	if err := ValidateContract(contract); err != nil {
		return SemanticIR{}, Contract{}, err
	}
	contractDigest := DigestBytes(source)
	ir := SemanticIR{
		Schema: "gooo/bounded-repair-proposal-projector/ir/v1", SourcePath: sourcePath, SourceDigest: contractDigest, ContractDigest: contractDigest,
		Total: contract.Total, ResolutionPrecedence: append([]State(nil), contract.ResolutionPrecedence...),
		Activities: append([]SourceActivity(nil), contract.Activities...), Cases: append([]ContractCase(nil), contract.Cases...),
	}
	return ir, contract, nil
}

func ValidateContract(contract Contract) error {
	if contract.Schema != SourceSchema || contract.ID == "" || contract.Total != 12 || len(contract.Activities) != 12 || len(contract.Cases) != 12 {
		return errors.New("INVALID_FIXED_DENOMINATOR")
	}
	if contract.ClosureState != "FIXED_POINT_ONLY" || !contract.NoAggregateScore || !contract.NoNaturalLanguageInference || contract.RuntimeApplyAuthority != 0 {
		return errors.New("FORBIDDEN_CLOSURE_OR_AUTHORITY_RULE")
	}
	if !sameStrings(contract.UnknownFields, []string{"stage", "step", "reason", "unknown_class", "next_operation", "blocked_by"}) {
		return errors.New("UNKNOWN_FIELDS_NOT_EXACTLY_SIX")
	}
	if !sameStates(contract.ResolutionPrecedence, []State{StateRefuted, StateUnknown, StateClosed}) {
		return errors.New("RESOLUTION_PRECEDENCE_MISMATCH")
	}
	if !sameStrings(contract.AuthoritySeparation, []string{"proposal_generation", "apply", "commit", "merge"}) {
		return errors.New("AUTHORITY_SEPARATION_MISMATCH")
	}
	proofCounts := map[string]int{}
	indicatorCounts := map[string]int{}
	seenActivity := map[string]bool{}
	for index, activity := range contract.Activities {
		if activity.ID == "" || activity.Activity == "" || activity.ClaimID == "" || activity.OperationID == "" || activity.Stage == "" || activity.Step == "" || seenActivity[activity.ID] || !validProof(activity.ProofChoice) || !validIndicator(activity.IndicatorClass) {
			return fmt.Errorf("INVALID_ACTIVITY_%d", index+1)
		}
		seenActivity[activity.ID] = true
		proofCounts[activity.ProofChoice]++
		indicatorCounts[activity.IndicatorClass]++
	}
	for _, value := range []string{"FOUNDATION", "COHERENCE", "REGRESSION"} {
		if proofCounts[value] != 4 {
			return fmt.Errorf("PROOF_VECTOR_NOT_EXACT_%s", value)
		}
	}
	for _, value := range []string{"DRIVER", "OUTCOME", "GUARDRAIL"} {
		if indicatorCounts[value] != 4 {
			return fmt.Errorf("INDICATOR_VECTOR_NOT_EXACT_%s", value)
		}
	}
	seenCase := map[string]bool{}
	for index, item := range contract.Cases {
		if item.Ordinal != index+1 || item.ID == "" || seenCase[item.ID] || (item.Class != "normal" && item.Class != "unknown" && item.Class != "refuted") || !validState(item.ExpectedState) || !validProof(item.ProofChoice) || !validIndicator(item.IndicatorClass) || len(item.Vector) != 12 {
			return fmt.Errorf("INVALID_CASE_%d", index+1)
		}
		seenCase[item.ID] = true
	}
	return nil
}

func parseAttributes(input string) (map[string]string, error) {
	attrs := map[string]string{}
	for position := 0; position < len(input); {
		for position < len(input) && input[position] == ' ' {
			position++
		}
		if position == len(input) {
			break
		}
		keyStart := position
		for position < len(input) && input[position] != '=' && input[position] != ' ' {
			position++
		}
		if keyStart == position || position >= len(input) || input[position] != '=' {
			return nil, errors.New("INVALID_ATTRIBUTE")
		}
		key := input[keyStart:position]
		if _, exists := attrs[key]; exists {
			return nil, errors.New("DUPLICATE_ATTRIBUTE")
		}
		position++
		if position >= len(input) || input[position] != '"' {
			return nil, errors.New("ATTRIBUTES_MUST_BE_QUOTED")
		}
		valueStart := position
		position++
		for position < len(input) {
			if input[position] == '"' && input[position-1] != '\\' {
				position++
				break
			}
			position++
		}
		if position == 0 || position > len(input) || input[position-1] != '"' {
			return nil, errors.New("UNTERMINATED_ATTRIBUTE")
		}
		value, err := strconv.Unquote(input[valueStart:position])
		if err != nil {
			return nil, err
		}
		attrs[key] = value
	}
	return attrs, nil
}

func splitComma(value string) []string {
	if value == "" {
		return nil
	}
	items := strings.Split(value, ",")
	for index := range items {
		items[index] = strings.TrimSpace(items[index])
	}
	return items
}

func parseVector(value string) []int {
	parts := splitComma(value)
	result := make([]int, len(parts))
	for index, item := range parts {
		result[index] = atoi(item)
	}
	return result
}

func atoi(value string) int {
	parsed, _ := strconv.Atoi(value)
	return parsed
}

func sameStrings(left, right []string) bool {
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

func sameStates(left, right []State) bool {
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

func validProof(value string) bool {
	return value == "FOUNDATION" || value == "COHERENCE" || value == "REGRESSION"
}

func validIndicator(value string) bool {
	return value == "DRIVER" || value == "OUTCOME" || value == "GUARDRAIL"
}

func validState(value State) bool {
	return value == StateClosed || value == StateUnknown || value == StateRefuted
}
