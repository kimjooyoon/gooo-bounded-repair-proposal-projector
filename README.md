# gooo-bounded-repair-proposal-projector

This repository is the bounded semantic authority for projecting a repair
proposal dossier from immutable Gooo evidence. It never applies code and never
mutates the input repository.

The input is an immutable semantic graph, failed or refuted conformance
evidence, an optional UNKNOWN record with exactly six preserved fields, and an
optional counterexample. A proposal is emitted only when typed semantic IDs,
operation declarations, exact before/after cell digests, preconditions,
unchanged boundaries, expected evidence, validation steps, and capability/effect
budgets all bind to the same graph and evidence identity. A title or natural
language statement cannot create a proposal.

Resolution is `REFUTED > UNKNOWN > CLOSED`. `CLOSED` requires an explicit
`FIXED_POINT` terminal state. Every UNKNOWN preserves exactly:
`stage`, `step`, `reason`, `unknown_class`, `next_operation`, and `blocked_by`.
Proof choices (`FOUNDATION`, `COHERENCE`, `REGRESSION`) and indicators
(`DRIVER`, `OUTCOME`, `GUARDRAIL`) are independent exact vectors. No aggregate
score or percentage is produced. Improvement is `null` with state `UNKNOWN`
unless the before/after pair has exact scenario, graph, toolchain, runner, and
metric identity.

`.gooo/repair-proposal-projector.gooo` owns the proposal semantics and the
twelve meta activities. Go 1.27 is only the parser, generated evaluator,
runtime, and artifact generator.

## Caller-owned projection

Run the projector from GitHub Actions or another controlled caller directory:

```text
go run ./cmd/gooo-repair-proposal-projector project \
  --input fixtures/inputs/normal-typed-repair.json \
  --output-dir /absolute/caller-owned/output
```

The output contains `proposal.json`, `proposal-events.ndjson`,
`semantic-ir.json`, `generated-evaluator.go`, `replay-receipt.json`, and
`human-report.md`. The output directory must be absolute, empty, and outside
the input repository.

The fixed conformance corpus has exactly twelve cases: four normal/CLOSED,
four UNKNOWN, and four REFUTED. Its exact vectors and preserved incidents are
bound in `fixtures/corpus.json` and the `.gooo` source.

## Authority boundary

Proposal generation authority is `1`. Apply, commit, merge, runtime apply,
repository-write, and release-mutation authorities are each `0`. The forbidden
operation set is explicit and includes apply, input-repository write, commit,
merge, release publication, and deletion. An accidental execution is retained
as `OPERATIONAL_REFUTED`.

The runtime writes only caller-owned output artifacts. It does not call GitHub,
edit a source graph, apply a proposal, create a commit, merge a branch, or
delete an asset.

## Validation and release

GitHub Actions is the validation boundary. Pull requests run Go 1.27
format/build/test/vet/conformance checks and actionlint; failed-run artifacts
are retained. The release workflow is manual and draft-first, requires a new
annotated tag, verifies exact asset SHA-256 digests and immutable release
metadata, and never deletes or reuses a tag, release, asset, or version.
