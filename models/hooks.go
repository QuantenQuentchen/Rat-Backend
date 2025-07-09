package models

type HookType int

const (
	HookVoteConcluded HookType = iota
	HookVoteCreated
	HookVoteSucceeded
	HookVoteFailed
	HookVoteVetoed
)
