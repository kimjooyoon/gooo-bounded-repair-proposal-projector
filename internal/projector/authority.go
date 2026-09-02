package projector

const (
	proposalOperationKind = "DECLARE_REPAIR_PROPOSAL"
	
	forbiddenApply    = "APPLY_CODE"
	forbiddenWrite    = "WRITE_INPUT_REPOSITORY"
	forbiddenCommit   = "COMMIT"
	forbiddenMerge    = "MERGE"
	forbiddenRelease  = "PUBLISH_RELEASE"
	forbiddenDelete   = "DELETE_ARTIFACT"
)

func fixedAuthority() AuthorityBoundary {
	return AuthorityBoundary{
		ProposalGenerationAuthority: 1,
		ApplyAuthority: 0,
		CommitAuthority: 0,
		MergeAuthority: 0,
		RuntimeApplyAuthority: 0,
		RepositoryWriteAuthority: 0,
		ReleaseMutationAuthority: 0,
	}
}

func fixedForbiddenOperations() []Operation {
	return []Operation{
		{ID: "forbidden.apply", Kind: forbiddenApply, Effect: "NEVER_EXECUTE"},
		{ID: "forbidden.repository-write", Kind: forbiddenWrite, Effect: "NEVER_EXECUTE"},
		{ID: "forbidden.commit", Kind: forbiddenCommit, Effect: "NEVER_EXECUTE"},
		{ID: "forbidden.merge", Kind: forbiddenMerge, Effect: "NEVER_EXECUTE"},
		{ID: "forbidden.release", Kind: forbiddenRelease, Effect: "NEVER_EXECUTE"},
		{ID: "forbidden.delete", Kind: forbiddenDelete, Effect: "NEVER_EXECUTE"},
	}
}

func forbiddenKind(value string) bool {
	switch value {
	case forbiddenApply, forbiddenWrite, forbiddenCommit, forbiddenMerge, forbiddenRelease, forbiddenDelete:
		return true
	default:
		return false
	}
}

func authorityZero(authority AuthorityBoundary) bool {
	return authority.ApplyAuthority == 0 && authority.CommitAuthority == 0 && authority.MergeAuthority == 0 && authority.RuntimeApplyAuthority == 0 && authority.RepositoryWriteAuthority == 0 && authority.ReleaseMutationAuthority == 0
}
