package db

import (
	"RatBackend/models"
	"fmt"
)

var roleArray = []models.Role{
	{
		Name: "Chairmen",
		Permissions: uint64(
			models.PermVote | models.PermStartVote | models.PermConcludeVote |
				models.PermVeto | models.PermInternalInfo | models.PermSuggest |
				models.PermCreateRole | models.PermAuditLog | models.PermContestVeto |
				models.PermSuspendUser | models.PermRemoveRole | models.PermAssignRole |
				models.PermRestoreUser),
		Unique: true,
	},
	{
		Name: "CoChairmen",
		Permissions: uint64(
			models.PermVote | models.PermStartVote | models.PermConcludeVote |
				models.PermVeto | models.PermInternalInfo | models.PermSuggest |
				models.PermCreateRole | models.PermAuditLog | models.PermContestVeto |
				models.PermSuspendUser | models.PermRemoveRole | models.PermAssignRole |
				models.PermRestoreUser),
		Unique:  true,
		Cascade: true,
	},
	{
		Name: "Member",
		Permissions: uint64(
			models.PermVote | models.PermPublicInfo | models.PermSuggest |
				models.PermContestVeto),
		Unique: false,
	},
	{
		Name:        "Observer",
		Permissions: uint64(models.PermPublicInfo | models.PermSuggest),
		Unique:      false,
	},
	{
		Name: "Watchdog",
		Permissions: uint64(
			models.PermPublicInfo | models.PermSuggest | models.PermInternalInfo |
				models.PermContestVeto | models.PermAuditLog | models.PermVeto),
		Unique: true,
	},
}

func createRoles() error {
	for _, role := range roleArray {
		_, err := AddRole(role)
		if err != nil {
			return fmt.Errorf("failed to insert role %s: %w", role.Name, err)
		}
	}
	return nil
}

func createSchemas() error {
	schema := `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS roles (
		    id BIGINT AUTOINCREMENT PRIMARY KEY,
			name TEXT PRIMARY KEY,
			perms BIGINT,
			unique_role BOOLEAN,
			timeout BIGINT,
			"cascade" BOOLEAN
		);
		CREATE TABLE IF NOT EXISTS role_bindings (
			user_id TEXT REFERENCES users(id),
			role_id TEXT REFERENCES roles(id),
			issuer_id TEXT REFERENCES users(id),
			issuer_role_id TEXT REFERENCES roles(id),
			issuedAt BIGINT,
			UNIQUE(user_id, role_id)
		);
		CREATE TABLE IF NOT EXISTS votes (
			id BIGINT AUTOINCREMENT PRIMARY KEY,
			name TEXT,
			description TEXT,
			kind INT,
			hasAbstain BOOLEAN,
			isPrivate BOOLEAN,
			voteState INT,
			timeout BIGINT,
			createdAt BIGINT
		);
		CREATE TABLE IF NOT EXISTS vote_bindings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			vote_id INT REFERENCES votes(id),
			user_id TEXT,
			choice INT,
			UNIQUE(vote_id, user_id)
		);
		CREATE TABLE IF NOT EXISTS role_assign_permissions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			assigner_role_id TEXT REFERENCES roles(id),
			assignee_role_id TEXT REFERENCES roles(id),
			UNIQUE(assigner_role_id, assignee_role_id)
		);
		CREATE TABLE IF NOT EXISTS vote_hooks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			vote_id INT REFERENCES votes(id),
			hook_type INT, --call hook type enum
			hook_data TEXT, -- JSON for args
			hook_function TEXT, -- function name to call
			UNIQUE(vote_id, hook_type)
			)
	`
	_, err := db.Exec(schema)
	return err
}
