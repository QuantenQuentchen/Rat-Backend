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

func InfoHandler(w http.ResponseWriter, r *http.Request) {
	authStruct, err := auth.Authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	switch r.URL.Path {
	case "/info/roles/all":
		db.GetAllRoles()

	case "/info/roles/user":
		userId := r.FormValue("user_id")
		db.GetUserRoles(userId)

	case "/info/create-vote":
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
	default:
		http.NotFound(w, r)
	}
}
