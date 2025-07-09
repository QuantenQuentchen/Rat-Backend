package db

import (
	"RatBackend/models"
	"database/sql"
	"fmt"
	"time"
)

func AddRole(role models.Role) (string, error) {
	result, err := db.NamedExec(`
		INSERT INTO roles (name, perms, unique_role, timeout, "cascade")
		VALUES (:name, :perms, :unique_role, :timeout, :cascade)`,
		role)
	if err != nil {
		return "", fmt.Errorf("failed to add role: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("failed to get last insert id: %w", err)
	}
	return fmt.Sprintf("%d", id), nil
}

func UpdateRole(roleId string, role models.Role) error {
	params := map[string]interface{}{
		"id":          roleId,
		"name":        role.Name,
		"perms":       role.Permissions,
		"unique_role": role.Unique,
		"timeout":     role.Timeout,
		"cascade":     role.Cascade,
	}
	_, err := db.NamedExec(`
		UPDATE roles
		SET name = :name, perms = :perms, unique_role = :unique_role, timeout = :timeout, "cascade" = :cascade
		WHERE id = :id`, params)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}
	return nil
}

func ValidateRole(role models.Role) error {
	if role.Timeout <= 0 {
		return fmt.Errorf("role timeout cannot be negative, or Null")
	}
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	_, err = tx.NamedExec(`
		INSERT INTO roles (name, perms, unique_role, timeout, "cascade")
		VALUES (:name, :perms, :unique_role, :timeout, :cascade)`,
		role)
	if err != nil {
		RBerr := tx.Rollback()
		return fmt.Errorf("failed to insert role: %w, %w", err, RBerr)
	}
	tx.Rollback()
	return nil
}

func GetAllRoles() ([]models.Role, error) {
	var roles []models.Role
	err := db.Select(&roles, ` SELECT * FROM roles`)
	if err != nil {
		return nil, err
	}
	return roles, err
}

func GetUserRoles(userID string) ([]models.Role, error) {
	var roles []models.Role
	err := db.Select(&roles, `
        SELECT r.id ,r.name, r.perms, r.unique_role, r.timeout, r."cascade"
        FROM role_bindings rb
        JOIN roles r ON rb.role_id = r.id
        WHERE rb.user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func GetRolesToRemove() ([]models.RoleBinding, error) {
	var roles []models.RoleBinding
	err := db.Select(&roles, `
        SELECT rb.user_id, rb.role_id, rb.issuer_id, rb.issuer_role_id, rb.issuedAt
        FROM role_bindings rb
        JOIN roles r ON rb.role_id = r.id
        WHERE rb.issuedAt + r.timeout <= ?`, time.Now().Unix())
	return roles, err
}

func IsCascadingRole(roleID string) (bool, error) {
	var isCascading bool
	err := db.Get(&isCascading, `
		SELECT "cascade" FROM roles WHERE id = ?
	`, roleID)
	if err != nil {
		return false, err
	}
	return isCascading, nil
}

// TODO: Add Logic for this
func GetUserIssuedBindings(userID string) ([]models.RoleBinding, error) {
	var bindings []models.RoleBinding
	err := db.Select(&bindings, `
		SELECT * FROM role_bindings WHERE issuer_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	return bindings, nil
}

func GetUserRoleIssuedBindings(userID string, roleID string) ([]models.RoleBinding, error) {
	var bindings []models.RoleBinding
	err := db.Select(&bindings, `
		SELECT * FROM role_bindings WHERE issuer_id = ? AND issuer_role_id = ?
	`, userID, roleID)
	if err != nil {
		return nil, err
	}
	return bindings, nil
}

func InsertVote(vote models.Vote) (uint64, error) {
	result, err := db.Exec(`
		INSERT INTO votes
		(name, description, kind, hasAbstain, isPrivate, voteState, timeout, createdAt)
		VALUES 
		(:name, :description, :kind, :hasAbstain, :isPrivate, :voteState, :timeout, :createdAt)`,
		vote)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func GetVote(id uint64) (*models.Vote, error) {
	var vote models.Vote
	err := db.Get(&vote, "SELECT * FROM votes WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &vote, nil
}

func IsOngoing(id uint64) (bool, error) {
	var isOngoing models.VoteState
	err := db.Get(&isOngoing, "SELECT voteState FROM votes WHERE id = ?", id)
	if err != nil {
		return false, err
	}
	return isOngoing == models.Ongoing, nil
}

func HasAbstain(id uint64) (bool, error) {
	var hasAbstain bool
	err := db.Get(&hasAbstain, "SELECT hasAbstain FROM votes WHERE id = ?", id)
	if err != nil {
		return false, err
	}
	return hasAbstain, nil
}

func AddOrChangeVote(user_id string, vote_id uint64, position models.Position) error {
	_, err := db.Exec(
		`INSERT INTO vote_bindings (vote_id, user_id, choice) 
		VALUES (?, ?, ?) 
		ON CONFLICT (vote_id, user_id) DO UPDATE SET choice =excluded.choice`,
		vote_id, user_id, position,
	)
	return err
}

func TallyVotes(voteID uint64) (map[models.Position]int, error) {
	rows, err := db.Queryx(
		`SELECT choice, COUNT(*) as count FROM vote_bindings WHERE vote_id = ? GROUP BY choice`,
		voteID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[models.Position]int)
	for rows.Next() {
		var choice int
		var count int
		if err := rows.Scan(&choice, &count); err != nil {
			return nil, err
		}
		result[models.Position(choice)] = count
	}
	return result, nil
}

func TallyVotesPersonal(voteID uint64) (map[models.Position][]string, error) {
	rows, err := db.Queryx(
		`SELECT choice, user_id FROM vote_bindings WHERE vote_id = ?`,
		voteID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[models.Position][]string)
	for rows.Next() {
		var choice int
		var userID string
		if err := rows.Scan(&choice, &userID); err != nil {
			return nil, err
		}
		pos := models.Position(choice)
		result[pos] = append(result[pos], userID)
	}
	return result, nil
}

func IsPrivate(voteID uint64) bool {
	var isPrivate bool
	err := db.Get(&isPrivate, `SELECT isPrivate FROM votes WHERE vote_id = ?`, voteID)
	if err != nil {
		return false
	}
	return isPrivate
}

func GetRole(roleID string) (*models.Role, error) {
	var role models.Role
	err := db.Get(&role, `
		SELECT id, name, perms, unique_role, timeout, "cascade"
		FROM roles
		WHERE id = ?`, roleID)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func GetRolePermFlags(roleID string) (models.PermFlag, error) {
	var perms models.PermFlag
	err := db.Select(&perms, `
			SELECT r.perms
			FROM roles r
			WHERE r.id = ?`, roleID)
	if err != nil {
		return 0, err
	}
	return perms, nil
}

func GetUserRolePermFlags(userID string) ([]models.PermFlag, error) {
	var perms []models.PermFlag
	err := db.Select(&perms, `
			SELECT r.perms
			FROM role_bindings rb
			JOIN roles r ON rb.role_id = r.id
			WHERE rb.user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	return perms, nil
}

func UserHasPermission(userID string, check models.PermFlag) (bool, error) {
	var perms []models.PermFlag
	err := db.Select(&perms, `
        SELECT r.perms
        FROM role_bindings rb
        JOIN roles r ON rb.role_id = r.id
        WHERE rb.user_id = ?`, userID)
	if err != nil {
		return false, err
	}
	for _, p := range perms {
		if p.Has(check) {
			return true, nil
		}
	}
	return false, nil
}

func GetVoteKind(voteID uint64) (models.VoteKind, error) {
	var kind models.VoteKind
	err := db.Get(&kind, "SELECT kind FROM votes WHERE id = ?", voteID)
	if err != nil {
		return models.Simple, err // Default to Simple if not found
	}
	return kind, nil
}

func SetVotePrivate(voteID uint64) error {
	err, _ := db.Exec(
		`UPDATE votes SET isPrivate = ? WHERE id = ?`,
		true, voteID,
	)
	if err != nil {
		return fmt.Errorf("failed to set vote %d as private: %w", voteID, err)
	}
	return nil
}

func GetVotesToConclude() ([]models.Vote, error) {
	var voteArr []models.Vote
	err := db.Select(&voteArr, "SELECT * FROM votes WHERE timeout <= ? AND voteState == ?", time.Now().Unix(), models.Ongoing)
	return voteArr, err
}

func UpdateVoteState(vote_id uint64, state models.VoteState) (sql.Result, error) {
	return db.Exec(
		`UPDATE votes SET voteState = ? WHERE id = ?`,
		state, vote_id,
	)
}

func RemoveRoleBinding(userID, roleID string) error {
	_, err := db.Exec(
		`DELETE FROM role_bindings WHERE user_id = ? AND role_id = ?`,
		userID, roleID,
	)
	return err
}

func AssignRoles(binding models.RoleBinding) error {
	var isUnique bool
	err := db.Get(&isUnique, "SELECT unique_role FROM roles WHERE id = ?", binding.RoleId)
	if err != nil {
		return fmt.Errorf("failed to check role uniqueness: %w", err)
	}
	if isUnique && !binding.Transferal {
		return fmt.Errorf("cannot assign unique role")
	}
	if isUnique && binding.Transferal {
		err := RemoveRoleBinding(binding.UserId, binding.RoleId)
		if err != nil {
			return fmt.Errorf("failed to remove role binding: %w", err)
		}
	}
	_, err = db.NamedExec("INSERT OR REPLACE INTO role_bindings (user_id, role_id) VALUES (:user_id, :role_id)", binding)
	if err != nil {
		return fmt.Errorf("failed to bind new role")
	}
	return nil
}

func AddUser(userID, name string) error {
	_, err := db.Exec(
		"INSERT INTO users (id, name) VALUES (?, ?)",
		userID, name,
	)
	return err
}
