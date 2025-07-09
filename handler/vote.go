package handler

import (
	"RatBackend/auth"
	"RatBackend/db"
	"RatBackend/logic"
	"RatBackend/models"
	"fmt"
	"net/http"
	"strconv"
)

func VoteHandler(w http.ResponseWriter, r *http.Request) {
	authStruct, err := auth.Authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	switch r.URL.Path {
	case "/vote/vote":
		voteID, err := strconv.Atoi(r.FormValue("vote_id"))
		if err != nil {
			fmt.Println("Error while parsing Form Val, %w", err)
		}
		i, err := strconv.Atoi(r.FormValue("position"))
		if err != nil {
			i = 0
		}
		position := models.Position(i)
		logic.Vote(authStruct.Subject, uint64(voteID), position)
	case "/vote/veto":
		voteID, err := strconv.Atoi(r.FormValue("vote_id"))
		if err != nil {
			fmt.Println("Error while parsing Form Val, %w", err)
		}
		logic.Veto(authStruct.Subject, uint64(voteID), r.FormValue("reason"))
	case "/vote/create-vote":
		name := r.FormValue("name")
		description := r.FormValue("description")
		isPrivateInt, err := strconv.Atoi(r.FormValue("is_private"))
		if err != nil {
			fmt.Println("Error while parsing Form Val, %w", err)
		}
		isPrivate := isPrivateInt != 0
		kindInt, err := strconv.Atoi(r.FormValue("kind"))
		if err != nil {
			fmt.Println("Error while parsing Form Val, %w", err)
		}
		kind := models.VoteKind(kindInt)

		timeout, err := strconv.ParseUint(r.FormValue("timeout"), 10, 64)

		if err != nil {
			fmt.Println("Error while parsing Form Val, %w", err)
		}

		err = logic.CreateVote(name, description, kind, timeout, isPrivate, true, authStruct.Subject)

		if err != nil {
			fmt.Println("Error while executing Vote Creation, %w", err)
		}
	case "/vote/private-vote":
		voteID, err := strconv.Atoi(r.FormValue("vote_id"))
		if err != nil {
			fmt.Println("Error while parsing Form Val, %w", err)
		}
		err = db.SetVotePrivate(uint64(voteID))
		if err != nil {
			fmt.Println("Error while setting vote as private, %w", err)
		}
	default:
		http.NotFound(w, r)
	}
}
