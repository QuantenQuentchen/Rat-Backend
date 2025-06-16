package logic

import (
	events "RatBackend/Events"
	"RatBackend/db"
	"RatBackend/models"
	"fmt"
	"time"
)

func conclude(vote models.Vote) {
	voteMap, err := db.TallyVotes(vote.ID)
	if err != nil {
		fmt.Printf("Failed to tally vote %d: %v\n", vote.ID, err)
		return
	}
	if voteMap[models.For] == voteMap[models.Against] && vote.HasAbstain {
		newID, err := createVote(
			vote.Name, vote.Description,
			vote.Kind, vote.CreatedAt-uint64(time.Now().Unix()),
			vote.IsPrivate, false)
		if err != nil {
			fmt.Errorf("Error in Conclude Check: %w", err)
		}
		changeVoteState(vote.ID, models.FailedInconclusive, models.WithNewVoteID(newID))
	}
	total := voteMap[models.For] + voteMap[models.Against] + voteMap[models.Abstain]
	if models.VoteKindFunctions[vote.Kind](voteMap[models.For], voteMap[models.Against], total) {
		changeVoteState(vote.ID, models.Succeeded)
	} else {
		changeVoteState(vote.ID, models.FailedVotes)
	}
}

func RemoveRoleBinding(user_id, role_id string) error {
	isCascading, err := db.IsCascadingRole(role_id)
	//events.BroadcastRoleUpdate(user_id, role_id, models.RoleRemoved)
	if err != nil {
		return fmt.Errorf("failed to check if role %s is cascading: %w", role_id, err)
	}
	db.RemoveRoleBinding(user_id, role_id)
	if isCascading {
		rolesToRemove, err := db.GetUserRoleIssuedBindings(user_id, role_id)
		if err != nil {
			return fmt.Errorf("failed to get roles to remove for user %s and role %s: %w", user_id, role_id, err)
		}
		for _, role := range rolesToRemove {
			err := RemoveRoleBinding(role.Role_id, role.User_id)
			if err != nil {
				return fmt.Errorf("failed to remove cascading role binding for user %s and role %s: %w", role.User_id, role.Role_id, err)
			}
		}
	}
	return nil
}

func changeVoteState(vote_id uint64, state models.VoteState, opts ...models.VoteStateOption) error {
	ctx := &models.VoteStateChangeContext{}
	for _, opt := range opts {
		opt(ctx)
	}
	_, err := db.UpdateVoteState(vote_id, state)
	if err != nil {
		return err
	}

	events.BroadcastVoteStateChange(vote_id, state, ctx)
	return nil
}

func createVote(name, description string, kind models.VoteKind, timeout uint64, isPrivate, hasAbstain bool) (uint64, error) {
	vote := models.Vote{
		ID:          0,
		Name:        name,
		Description: description,
		Kind:        kind,
		HasAbstain:  hasAbstain,
		IsPrivate:   isPrivate,
		VoteState:   models.Ongoing,
		Timeout:     timeout,
		CreatedAt:   uint64(time.Now().Unix()),
	}
	id, err := db.InsertVote(vote)
	if err != nil {
		fmt.Errorf("Error creating vote %w", err)
	}
	vote.ID = id
	events.BroadcastVoteCreation(vote)
	return id, nil
}

func Vote(user_id string, vote_id uint64, position models.Position) error {
	allowed, err := db.UserHasPermission(user_id, models.PermVote)
	if err != nil {
		return err
	}
	if !allowed {
		return models.ErrMissingPermission
	}

	isOngoing, err := db.IsOngoing(vote_id)
	if err != nil {
		return err
	}
	if !isOngoing {
		return models.ErrNotOngoing
	}
	if position == models.Abstain {
		hasAbstain, err := db.HasAbstain(vote_id)
		if err != nil {
			return err
		}
		if !hasAbstain {
			return models.ErrNoAbstain
		}
	}
	err = db.AddOrChangeVote(user_id, vote_id, position)
	if err != nil {
		return err
	}

	return nil
}

func Veto(user_id string, vote_id uint64, reason string) error {
	allowed, err := db.UserHasPermission(user_id, models.PermVeto)
	if err != nil {
		return err
	}
	if !allowed {
		return models.ErrMissingPermission
	}
	isOngoing, err := db.IsOngoing(vote_id)
	if err != nil {
		return err
	}
	if !isOngoing {
		return models.ErrNotOngoing
	}
	return changeVoteState(vote_id, models.FailedVeto, models.WithVetoDetails(user_id, reason))
}

func CreateVote(name, description string, kind models.VoteKind, timeout uint64, isPrivate, hasAbstain bool, user_id string) error {
	allowed, err := db.UserHasPermission(user_id, models.PermStartVote)
	if err != nil {
		return err
	}
	if !allowed {
		return models.ErrMissingPermission
	}
	_, err = createVote(name, description, kind, timeout, isPrivate, hasAbstain)
	return err
}
