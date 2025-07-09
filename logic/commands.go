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
		err = changeVoteState(vote.ID, models.Succeeded)
		if err != nil {
			return
		}
	} else {
		err = changeVoteState(vote.ID, models.FailedVotes)
		if err != nil {
			return
		}
	}
}

func RemoveRoleBinding(userId, roleId string, ctx *models.RoleChangeContext, opts ...models.RoleChangeOption) error {
	isCascading, err := db.IsCascadingRole(roleId)
	if ctx == nil {
		ctx = &models.RoleChangeContext{}
	}
	for _, opt := range opts {
		opt(ctx)
	}
	events.BroadcastRoleUpdate(userId, roleId, events.RoleRemoved, ctx)
	if err != nil {
		return fmt.Errorf("failed to check if role %s is cascading: %w", roleId, err)
	}
	err = db.RemoveRoleBinding(userId, roleId)
	if err != nil {
		return err
	}
	if isCascading {
		var rolesToRemove, err = db.GetUserRoleIssuedBindings(userId, roleId)
		if err != nil {
			return fmt.Errorf("failed to get roles to remove for user %s and role %s: %w", userId, roleId, err)
		}
		for _, role := range rolesToRemove {
			err = RemoveRoleBinding(role.RoleId, role.UserId, ctx, models.WithCascadeReason())
			if err != nil {
				return fmt.Errorf("failed to remove cascading role binding for user %s and role %s: %w", role.UserId, role.RoleId, err)
			}
		}
	}
	return nil
}

func changeVoteState(voteId uint64, state models.VoteState, opts ...models.VoteStateOption) error {
	ctx := &models.VoteStateChangeContext{}
	for _, opt := range opts {
		opt(ctx)
	}
	_, err := db.UpdateVoteState(voteId, state)
	if err != nil {
		return err
	}

	events.BroadcastVoteStateChange(voteId, state, ctx)
	if state == models.Succeeded || state == models.SucceededEmergency {
		//somehow call an on_suceeded hook
	}
	return nil
}

func SetVotePrivate(voteId uint64) error {

	isOngoing, err := db.IsOngoing(voteId)
	if err != nil {
		return err
	}
	if !isOngoing {
		return models.ErrNotOngoing
	}
	kind, err := db.GetVoteKind(voteId)
	if err != nil {
		return fmt.Errorf("failed to get vote kind for vote %d: %w", voteId, err)
	}
	if kind != models.Simple {
		return fmt.Errorf("vote %d is not a simple vote, cannot set private", voteId)
	}
	err = db.SetVotePrivate(voteId)
	if err != nil {
		return err
	}

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
		fmt.Println("error creating vote %w", err)
	}
	vote.ID = id
	events.BroadcastVoteCreation(vote)
	return id, nil
}

func Vote(userId string, voteId uint64, position models.Position) error {
	allowed, err := db.UserHasPermission(userId, models.PermVote)
	if err != nil {
		return err
	}
	if !allowed {
		return models.ErrMissingPermission
	}

	isOngoing, err := db.IsOngoing(voteId)
	if err != nil {
		return err
	}
	if !isOngoing {
		return models.ErrNotOngoing
	}
	if position == models.Abstain {
		hasAbstain, err := db.HasAbstain(voteId)
		if err != nil {
			return err
		}
		if !hasAbstain {
			return models.ErrNoAbstain
		}
	}
	err = db.AddOrChangeVote(userId, voteId, position)
	if err != nil {
		return err
	}

	return nil
}

func Veto(userId string, voteId uint64, reason string) error {
	allowed, err := db.UserHasPermission(userId, models.PermVeto)
	if err != nil {
		return err
	}
	if !allowed {
		return models.ErrMissingPermission
	}
	isOngoing, err := db.IsOngoing(voteId)
	if err != nil {
		return err
	}
	if !isOngoing {
		return models.ErrNotOngoing
	}
	return changeVoteState(voteId, models.FailedVeto, models.WithVetoDetails(userId, reason))
}

func CreateVote(name, description string, kind models.VoteKind, timeout uint64, isPrivate, hasAbstain bool, userId string) error {
	allowed, err := db.UserHasPermission(userId, models.PermStartVote)
	if err != nil {
		return err
	}
	if !allowed {
		return models.ErrMissingPermission
	}
	_, err = createVote(name, description, kind, timeout, isPrivate, hasAbstain)
	return err
}
