package handler

import (
	"RatBackend/api"
	"RatBackend/auth"
	"RatBackend/db"
	_ "RatBackend/logic"
	"RatBackend/models"
	_ "RatBackend/models"
	"encoding/json"
	"fmt"
	"net/http"
	_ "strconv"
)

func InfoHandler(w http.ResponseWriter, r *http.Request) {
	authStruct, err := auth.Authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	switch r.URL.Path {

	case "info/permissions/get-all-permissions":
		response := map[string]interface{}{
			"status":        "ok",
			"permissionMap": models.PermFlagNames,
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
			return
		}

	case "/info/roles/all":
		allRoles, err := db.GetAllRoles()
		if err != nil {
			return
		}
		response := map[string]interface{}{
			"status": "ok",
			"roles":  allRoles,
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
			return
		}

	case "info/roles/get-role-permissions":
		roleId := r.FormValue("role_id")
		if roleId == "" {
			http.Error(w, "Role ID is required", http.StatusBadRequest)
			return
		}
		permissions, err := db.GetRolePermFlags(roleId)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error retrieving permissions: %v", err), http.StatusInternalServerError)
			return
		}
		response := map[string]interface{}{
			"status":      "ok",
			"role_id":     roleId,
			"permissions": permissions,
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
			return
		}
	// Assuming permissions is a slice of strings or a similar type
	case "info/roles/get-user-permissions":
		userId := r.FormValue("user_id")
		if userId == "" {
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}
		permissions, err := api.GetUserPermissions(userId)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error retrieving permissions: %v", err), http.StatusInternalServerError)
			return
		}
		response := map[string]interface{}{
			"status":      "ok",
			"user_id":     userId,
			"permissions": permissions,
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
			return
		}
	case "/inf/roles/get-user-roles":
		userId := r.FormValue("user_id")
		if userId == "" {
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}
		roles, err := api.GetUserRoles(userId)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error retrieving roles: %v", err), http.StatusInternalServerError)
			return
		}
		response := map[string]interface{}{
			"status":  "ok",
			"user_id": userId,
			"roles":   roles,
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
			return
		}
	case "/info/roles/get-role":
		roleId := r.FormValue("role_id")
		if roleId == "" {
			http.Error(w, "Role ID is required", http.StatusBadRequest)
			return
		}
		role, err := db.GetRole(roleId)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error retrieving role: %v", err), http.StatusInternalServerError)
			return
		}
		response := map[string]interface{}{
			"status": "ok",
			"role":   role,
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)

	default:
		http.NotFound(w, r)
	}
}
