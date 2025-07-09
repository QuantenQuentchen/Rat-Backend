package handler

import (
	"RatBackend/auth"
	"RatBackend/db"
	"RatBackend/models"
	"fmt"
	"net/http"
	"strconv"
)

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	authStruct, err := auth.Authenticate(r)
	if err != nil {
		fmt.Println("Auth Failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if authStruct.Issuer != "gov-bot" {
		http.Error(w, models.ErrReservedEndpoint.Error(), http.StatusUnauthorized)
		return
	}
	switch r.URL.Path {
	case "/admin/assign-role":
		isTransferalInt, err := strconv.Atoi(r.FormValue("is_transferal"))

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		binding := models.RoleBinding{
			UserId:     r.FormValue("user_id"),
			RoleId:     r.FormValue("role_id"),
			Transferal: isTransferalInt != 0,
		}
		err = db.AssignRoles(binding)
		if err != nil {
			return
		}
	case "/admin/create-user":
		user_id := authStruct.Subject
		name := r.FormValue("name")
		err := db.AddUser(user_id, name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	default:
		http.NotFound(w, r)
	}
}
