package api

import (
	"RatBackend/logic"
	"RatBackend/models"
)

type VoteCreationRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Kind        models.VoteKind `json:"kind"`
	Timeout     uint64          `json:"timeout"`
	IsPrivate   bool            `json:"is_private"`
	Identity    string          `json:"identity"`
}

type VoteVetoRequest struct {
	Identity string `json:"identity"`
	VoteID   uint64 `json:"vote_id"`
	Reason   string `json:"reason"`
}

type VotePrivateRequest struct {
	VoteID uint64 `json:"vote_id"`
}

func CreateVote(request VoteCreationRequest) error {
	name := request.Name
	description := request.Description
	kind := request.Kind
	timeout := request.Timeout
	isPrivate := request.IsPrivate
	identity := request.Identity
	err := logic.CreateVote(name, description, kind, timeout, isPrivate, true, identity)
	return err
}

func VetoVote(request VoteVetoRequest) error {
	identity := request.Identity
	voteID := request.VoteID
	reason := request.Reason
	err := logic.Veto(identity, voteID, reason)
	return err
}

func SetVotePrivate(request VotePrivateRequest) error {
	voteID := request.VoteID
	err := logic.SetVotePrivate(voteID)
	return err
}
