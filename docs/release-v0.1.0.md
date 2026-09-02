# Release v0.1.0

This is the first immutable patch release of the bounded repair proposal
projector. It contains the twelve-case conformance corpus, typed dossier
schema, exact proof/indicator vectors, and fail-closed authority boundary.

Release identity is an annotated tag bound to one commit. The draft-first
workflow uploads a source archive, evidence archive, manifest, and SHA256SUMS;
it verifies each asset digest and the immutable release API object before
publishing. Failed workflow runs and their artifacts are retained. Tags,
releases, assets, and versions are never deleted or reused.
