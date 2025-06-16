package db

import (
	"RatBackend/models"
	"fmt"
)

var roleArray = []models.Role{
	{
		Name:        "Chairmen",
		Permissions: uint64(models.PermVote | models.PermStartVote | models.PermConcludeVote | models.PermVeto | models.PermInternalInfo | models.PermSuggest),
		Unique:      true,
	},
	{
		Name:        "CoChairmen",
		Permissions: uint64(models.PermVote | models.PermStartVote | models.PermConcludeVote | models.PermVeto | models.PermInternalInfo | models.PermSuggest),
		Unique:      true,
	},
	{
		Name:        "Member",
		Permissions: uint64(models.PermVote | models.PermPublicInfo | models.PermSuggest),
		Unique:      false,
	},
	{
		Name:        "Observer",
		Permissions: uint64(models.PermPublicInfo | models.PermSuggest),
		Unique:      false,
	},
	{
		Name:        "Watchdog",
		Permissions: uint64(models.PermPublicInfo | models.PermSuggest | models.PermInternalInfo),
		Unique:      true,
	},
}

func createRoles() error {
	query := `INSERT OR REPLACE INTO roles (id, perms, unique_role) VALUES (:id, :perms, :unique_role)`
	for _, role := range roleArray {
		_, err := db.NamedExec(query, role)
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
			id TEXT PRIMARY KEY,
			perms BIGINT,
			unique_role BOOL,
			timeout BIGINT,
			cascade BOOL
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
			hasAbstain bool,
			isPrivate bool,
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
		)
	`
	_, err := db.Exec(schema)
	return err
}
