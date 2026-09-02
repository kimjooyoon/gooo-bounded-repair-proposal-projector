# Protocol v1

## Input

`schema` is `gooo/bounded-repair-proposal-projector/protocol/v1/input/v1`.
`semantic_graph` and `evidence` carry explicit immutable release and digest
identity. Evidence records are typed as `FAILED`, `REFUTED`, or `CLOSED`.
`FAILED` records may carry a `repair` seed. The seed is not code; it is a
typed declaration of the proposal boundary.

The optional `unknown` object is never normalized or summarized. Its exact
fields are `stage`, `step`, `reason`, `unknown_class`, `next_operation`, and
`blocked_by`. If it is present and valid, it is copied byte-for-value in the
semantic object model and emitted under `unresolved_frontier`.

The optional `counterexample` must bind its graph digest, failure evidence ID,
target IDs, observed value/digest, and expected value/digest. Identity mismatch
is `REFUTED`, not a new proposal.

## Proposal

`proposal.json` is `null` for UNKNOWN and REFUTED decisions. A CLOSED proposal
contains these typed fields:

- `target_semantic_ids`
- `allowed_operations`
- `forbidden_operations`
- `preconditions`
- `claimed_changed_cells`
- `unchanged_boundary`
- `expected_evidence`
- `validation_plan`
- `capability_effect_budget`
- `unresolved_frontier`
- `authority`

The changed-cell claim requires an exact graph-bound `before_digest` and a
different exact `after_digest`; it does not apply the change. The proposal
operation is `DECLARE_REPAIR_PROPOSAL` with effect `PROPOSAL_ONLY`. Direct
application and all repository/release mutations are forbidden.

## Resolution

The evaluator uses the fixed precedence `REFUTED > UNKNOWN > CLOSED`.
`CLOSED` sets `decision` to `FIXED_POINT` only when the evidence bundle itself
contains the literal terminal state `FIXED_POINT`. Any other terminal state is
UNKNOWN. This is a projector closure, not product adoption.

The exact denominator is twelve meta activities, with four each of
`FOUNDATION`, `COHERENCE`, and `REGRESSION`, independently four each of
`DRIVER`, `OUTCOME`, and `GUARDRAIL`. No aggregate score exists.

Improvement is evaluated separately. Without an exact matching before/after
pair for scenario, graph digest, toolchain, runner, and metric, the output is
`improvement: null` and `improvement_state: UNKNOWN`.

## Artifact and authority rules

The runtime accepts only an absolute empty output directory outside the input
repository. It produces proposal, event, semantic IR, generated evaluator,
replay, and human report artifacts. No artifact is written into the input
repository. Repository writes, apply, commit, merge, and release mutation are
fixed integer zero in the runtime receipt.
