package projector

const (
	ProtocolSchema = "gooo/bounded-repair-proposal-projector/protocol/v1"
	SourceSchema   = "gooo/bounded-repair-proposal-projector/v1"
	IRSchema       = "gooo/bounded-repair-proposal-projector/ir/v1"
	
	StateClosed  State = "CLOSED"
	StateUnknown State = "UNKNOWN"
	StateRefuted State = "REFUTED"
)

type State string

type ProofChoice string
type IndicatorClass string

type GraphIdentity struct {
	Repository string `json:"repository"`
	ReleaseID  string `json:"release_id"`
	Tag        string `json:"tag"`
	CommitSHA  string `json:"commit_sha"`
	Asset      string `json:"asset"`
	AssetDigest string `json:"asset_digest"`
}

type SemanticGraph struct {
	Schema    string         `json:"schema"`
	Identity  GraphIdentity  `json:"identity"`
	Immutable bool           `json:"immutable"`
	Digest    string         `json:"digest"`
	Nodes     []SemanticNode `json:"nodes"`
	Edges     []SemanticEdge `json:"edges"`
}

type SemanticNode struct {
	ID             string `json:"id"`
	Kind           string `json:"kind"`
	Digest         string `json:"digest"`
	ContractCellID string `json:"contract_cell_id"`
}

type SemanticEdge struct {
	ID   string `json:"id"`
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

type EvidenceBundle struct {
	Schema        string           `json:"schema"`
	Immutable     bool             `json:"immutable"`
	Digest        string           `json:"digest"`
	GraphDigest   string           `json:"graph_digest"`
	TerminalState string           `json:"terminal_state"`
	Records       []EvidenceRecord `json:"records"`
}

type EvidenceRecord struct {
	ID                  string              `json:"id"`
	Status              string              `json:"status"`
	Immutable           bool                `json:"immutable"`
	GraphDigest         string              `json:"graph_digest"`
	ProofChoice         string              `json:"proof_choice"`
	IndicatorClass      string              `json:"indicator_class"`
	TargetSemanticIDs   []string            `json:"target_semantic_ids"`
	CausalEdgeIDs       []string            `json:"causal_edge_ids"`
	Repair              *RepairSeed         `json:"repair"`
	CounterexampleID    string              `json:"counterexample_id"`
	Unknown             *UnknownRecord      `json:"unknown"`
	Reason              string              `json:"reason"`
}

type UnknownRecord struct {
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	UnknownClass  string `json:"unknown_class"`
	NextOperation string `json:"next_operation"`
	BlockedBy     string `json:"blocked_by"`
}

type Counterexample struct {
	ID                string   `json:"id"`
	GraphDigest       string   `json:"graph_digest"`
	FailureEvidenceID string   `json:"failure_evidence_id"`
	TargetSemanticIDs []string `json:"target_semantic_ids"`
	ObservedValue     string   `json:"observed_value"`
	ExpectedValue     string   `json:"expected_value"`
	ObservedDigest    string   `json:"observed_digest"`
	ExpectedDigest    string   `json:"expected_digest"`
}

type RepairSeed struct {
	TargetSemanticIDs      []string                 `json:"target_semantic_ids"`
	AllowedOperations      []Operation              `json:"allowed_operations"`
	ForbiddenOperations    []Operation              `json:"forbidden_operations"`
	Preconditions          []Precondition            `json:"preconditions"`
	ClaimedChangedCells    []ChangedCell             `json:"claimed_changed_cells"`
	UnchangedBoundary      []BoundaryClaim           `json:"unchanged_boundary"`
	ExpectedEvidence       []EvidenceExpectation     `json:"expected_evidence"`
	ValidationPlan         []ValidationStep          `json:"validation_plan"`
	CapabilityEffectBudget CapabilityEffectBudget    `json:"capability_effect_budget"`
}

type Operation struct {
	ID                string   `json:"id"`
	Kind              string   `json:"kind"`
	TargetSemanticIDs []string `json:"target_semantic_ids"`
	Effect            string   `json:"effect"`
}

type Precondition struct {
	ID                string   `json:"id"`
	Kind              string   `json:"kind"`
	TargetSemanticIDs []string `json:"target_semantic_ids"`
	RequiredValue     string   `json:"required_value"`
}

type ChangedCell struct {
	SemanticID     string `json:"semantic_id"`
	ContractCellID string `json:"contract_cell_id"`
	BeforeDigest   string `json:"before_digest"`
	AfterDigest    string `json:"after_digest"`
	EvidenceID     string `json:"evidence_id"`
}

type BoundaryClaim struct {
	ID                string   `json:"id"`
	Kind              string   `json:"kind"`
	SemanticIDs       []string `json:"semantic_ids"`
	MustRemainUnchanged bool   `json:"must_remain_unchanged"`
}

type EvidenceExpectation struct {
	ID                string   `json:"id"`
	Kind              string   `json:"kind"`
	TargetSemanticIDs []string `json:"target_semantic_ids"`
	RequiredStatus    string   `json:"required_status"`
	ReceiptFields     []string `json:"receipt_fields"`
}

type ValidationStep struct {
	ID                string   `json:"id"`
	Kind              string   `json:"kind"`
	TargetSemanticIDs []string `json:"target_semantic_ids"`
	RequiredEvidence  []string `json:"required_evidence"`
	Authority         string   `json:"authority"`
}

type CapabilityEffectBudget struct {
	Capabilities []CapabilityBudget `json:"capabilities"`
	Effects      []EffectBudget     `json:"effects"`
}

type CapabilityBudget struct {
	ID       string `json:"id"`
	Limit    int    `json:"limit"`
	MayApply bool   `json:"may_apply"`
}

type EffectBudget struct {
	ID       string `json:"id"`
	Limit    int    `json:"limit"`
	MayApply bool   `json:"may_apply"`
}

type OperationalEvent struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Executed bool   `json:"executed"`
	Target   string `json:"target"`
}

type UtilityPair struct {
	ScenarioID string         `json:"scenario_id"`
	GraphDigest string        `json:"graph_digest"`
	Toolchain  string         `json:"toolchain"`
	Runner     string         `json:"runner"`
	MetricID   string         `json:"metric_id"`
	Before     *UtilityVector `json:"before"`
	After      *UtilityVector `json:"after"`
}

type UtilityVector struct {
	WallMS   int `json:"wall_ms"`
	RSSKib   int `json:"rss_kib"`
}

type Input struct {
	Schema            string             `json:"schema"`
	CaseID            string             `json:"case_id"`
	Graph             SemanticGraph      `json:"semantic_graph"`
	Evidence          EvidenceBundle     `json:"evidence"`
	Unknown           *UnknownRecord     `json:"unknown"`
	Counterexample    *Counterexample    `json:"counterexample"`
	Utility           *UtilityPair       `json:"utility"`
	OperationalEvents []OperationalEvent `json:"operational_events"`
}

type Contract struct {
	Schema                    string         `json:"schema"`
	ID                        string         `json:"id"`
	Total                     int            `json:"total"`
	ResolutionPrecedence      []State        `json:"resolution_precedence"`
	ClosureState              string         `json:"closure_state"`
	UnknownFields             []string       `json:"unknown_fields"`
	NoAggregateScore          bool           `json:"no_aggregate_score"`
	NoNaturalLanguageInference bool          `json:"no_natural_language_inference"`
	RuntimeApplyAuthority     int            `json:"runtime_apply_authority"`
	AuthoritySeparation       []string       `json:"authority_separation"`
	Activities                []SourceActivity `json:"activities"`
	Cases                     []ContractCase  `json:"cases"`
}

type SourceActivity struct {
	ID             string `json:"id"`
	Activity       string `json:"activity"`
	ClaimID        string `json:"claim_id"`
	OperationID    string `json:"operation_id"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	ProofChoice    string `json:"proof_choice"`
	IndicatorClass string `json:"indicator_class"`
	SourceLine     int    `json:"source_line"`
}

type ContractCase struct {
	Ordinal        int    `json:"ordinal"`
	ID             string `json:"id"`
	Class          string `json:"class"`
	ExpectedState  State  `json:"expected_state"`
	ProofChoice    string `json:"proof_choice"`
	IndicatorClass string `json:"indicator_class"`
	Vector         []int  `json:"vector"`
}

type SemanticIR struct {
	Schema              string          `json:"schema"`
	SourcePath          string          `json:"source_path"`
	SourceDigest        string          `json:"source_digest"`
	ContractDigest      string          `json:"contract_digest"`
	Total               int             `json:"total"`
	ResolutionPrecedence []State         `json:"resolution_precedence"`
	Activities          []SourceActivity `json:"activities"`
	Cases               []ContractCase   `json:"cases"`
}

type AuthorityBoundary struct {
	ProposalGenerationAuthority int `json:"proposal_generation_authority"`
	ApplyAuthority              int `json:"apply_authority"`
	CommitAuthority             int `json:"commit_authority"`
	MergeAuthority              int `json:"merge_authority"`
	RuntimeApplyAuthority       int `json:"runtime_apply_authority"`
	RepositoryWriteAuthority    int `json:"repository_write_authority"`
	ReleaseMutationAuthority    int `json:"release_mutation_authority"`
}

type RepairProposal struct {
	Schema                  string                  `json:"schema"`
	ProposalID              string                  `json:"proposal_id"`
	SourceEvidenceID        string                  `json:"source_evidence_id"`
	SourceGraphDigest       string                  `json:"source_graph_digest"`
	TargetSemanticIDs       []string                `json:"target_semantic_ids"`
	AllowedOperations       []Operation              `json:"allowed_operations"`
	ForbiddenOperations     []Operation              `json:"forbidden_operations"`
	Preconditions            []Precondition           `json:"preconditions"`
	ClaimedChangedCells      []ChangedCell            `json:"claimed_changed_cells"`
	UnchangedBoundary        []BoundaryClaim          `json:"unchanged_boundary"`
	ExpectedEvidence         []EvidenceExpectation    `json:"expected_evidence"`
	ValidationPlan           []ValidationStep         `json:"validation_plan"`
	CapabilityEffectBudget   CapabilityEffectBudget   `json:"capability_effect_budget"`
	UnresolvedFrontier       []FrontierItem            `json:"unresolved_frontier"`
	Authority                 AuthorityBoundary        `json:"authority"`
}

type FrontierItem struct {
	ID                string        `json:"frontier_id"`
	Kind              string        `json:"kind"`
	TargetSemanticIDs []string      `json:"target_semantic_ids"`
	Unknown           UnknownRecord `json:"unknown"`
}

type RefutedIncident struct {
	ID                string   `json:"incident_id"`
	Kind              string   `json:"kind"`
	EvidenceID        string   `json:"evidence_id"`
	TargetSemanticIDs []string `json:"target_semantic_ids"`
	Operation         string   `json:"operation"`
	Reason            string   `json:"reason"`
}

type ExactVector struct {
	TargetSemanticIDs    int `json:"target_semantic_ids"`
	AllowedOperations    int `json:"allowed_operations"`
	ForbiddenOperations  int `json:"forbidden_operations"`
	Preconditions        int `json:"preconditions"`
	ClaimedChangedCells  int `json:"claimed_changed_cells"`
	UnchangedBoundary    int `json:"unchanged_boundary"`
	ExpectedEvidence     int `json:"expected_evidence"`
	ValidationPlan       int `json:"validation_plan"`
	CapabilityBudget     int `json:"capability_budget"`
	EffectBudget         int `json:"effect_budget"`
	UnresolvedFrontier   int `json:"unresolved_frontier"`
	Incidents            int `json:"incidents"`
}

type BucketVector struct {
	FOUNDATION  int `json:"FOUNDATION"`
	COHERENCE   int `json:"COHERENCE"`
	REGRESSION  int `json:"REGRESSION"`
}

type IndicatorVector struct {
	DRIVER    int `json:"DRIVER"`
	OUTCOME   int `json:"OUTCOME"`
	GUARDRAIL int `json:"GUARDRAIL"`
}

type ImprovementClaim struct {
	MetricID string        `json:"metric_id"`
	Before   UtilityVector `json:"before"`
	After    UtilityVector `json:"after"`
	Delta    UtilityVector `json:"delta"`
}

type Dossier struct {
	Schema               string             `json:"schema"`
	CaseID               string             `json:"case_id"`
	InputDigest          string             `json:"input_digest"`
	GraphDigest          string             `json:"graph_digest"`
	State                State              `json:"state"`
	Decision             string             `json:"decision"`
	ResolutionPrecedence []State            `json:"resolution_precedence"`
	Proposal             *RepairProposal    `json:"proposal"`
	UnknownFrontier      []FrontierItem     `json:"unresolved_frontier"`
	RefutedIncidents     []RefutedIncident  `json:"refuted_incidents"`
	ExactVector          ExactVector        `json:"exact_vector"`
	ProofVector          BucketVector       `json:"proof_vector"`
	IndicatorVector      IndicatorVector   `json:"indicator_vector"`
	ImprovementState     State              `json:"improvement_state"`
	Improvement          *ImprovementClaim  `json:"improvement"`
	Authority             AuthorityBoundary  `json:"authority"`
}

type ProposalEvent struct {
	Schema       string   `json:"schema"`
	Ordinal      int      `json:"ordinal"`
	EventID      string   `json:"event_id"`
	EventType    string   `json:"event_type"`
	CaseID       string   `json:"case_id"`
	EvidenceID   string   `json:"evidence_id"`
	TargetIDs    []string `json:"target_semantic_ids"`
	InputDigest  string   `json:"input_digest"`
	Reason       string   `json:"reason"`
	EventDigest  string   `json:"event_digest"`
}

type ArtifactBinding struct {
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

type ReplayReceipt struct {
	Schema              string            `json:"schema"`
	CaseID              string            `json:"case_id"`
	InputDigest         string            `json:"input_digest"`
	Deterministic       bool              `json:"deterministic"`
	ProposalDigest      string            `json:"proposal_digest"`
	EventsDigest        string            `json:"events_digest"`
	OutputArtifacts     []ArtifactBinding `json:"output_artifacts"`
	RepositoryWrites    int               `json:"repository_writes"`
	ApplyExecutions     int               `json:"apply_executions"`
	CommitExecutions    int               `json:"commit_executions"`
	MergeExecutions     int               `json:"merge_executions"`
}

type Evaluation struct {
	Dossier         Dossier
	Events          []ProposalEvent
	SemanticIRRaw   []byte
	GeneratedRaw    []byte
	Replay          ReplayReceipt
}

type Corpus struct {
	Schema        string        `json:"schema"`
	CorpusID      string        `json:"corpus_id"`
	DenominatorID string        `json:"denominator_id"`
	Total         int           `json:"total"`
	Cases         []CorpusCase  `json:"cases"`
}

type CorpusCase struct {
	Ordinal           int      `json:"ordinal"`
	CaseID            string   `json:"case_id"`
	Class             string   `json:"class"`
	Input             string   `json:"input"`
	ExpectedState     State    `json:"expected_state"`
	ExpectedVector    []int    `json:"expected_vector"`
	ProofChoice       string   `json:"proof_choice"`
	IndicatorClass    string   `json:"indicator_class"`
	ExpectedIncidentIDs []string `json:"expected_incident_ids"`
}

type ConformanceResult struct {
	Ordinal        int          `json:"ordinal"`
	CaseID         string       `json:"case_id"`
	Class          string       `json:"class"`
	State          State        `json:"state"`
	ExactVector    ExactVector  `json:"exact_vector"`
	ProofChoice    string       `json:"proof_choice"`
	IndicatorClass string       `json:"indicator_class"`
	IncidentIDs    []string     `json:"incident_ids"`
	OutputFiles    int          `json:"output_files"`
}

type ConformanceIndex struct {
	Schema        string              `json:"schema"`
	CorpusID      string              `json:"corpus_id"`
	DenominatorID string              `json:"denominator_id"`
	Total         int                 `json:"total"`
	States        map[string]int      `json:"states"`
	Results       []ConformanceResult `json:"results"`
}
