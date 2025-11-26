package repository

import (
	"database/sql"
	"fmt"
	"strings"
)

type ActorRoleMappingRepository struct {
	DB *sql.DB
}

func NewActorRoleMappingRepository(db *sql.DB) *ActorRoleMappingRepository {
	return &ActorRoleMappingRepository{DB: db}
}

func (r *ActorRoleMappingRepository) GetRolesForUserContext(positionID int, contexts []string) ([]string, error) {
	if len(contexts) == 0 {
		return []string{}, nil
	}

	placeholders := make([]string, len(contexts))
	args := make([]interface{}, len(contexts)+1)
	args[0] = positionID
	for i, context := range contexts {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = context
	}

	query := fmt.Sprintf(`
        SELECT ar.name
        FROM actor_role_mapping arm
        JOIN actor_role ar ON arm.actor_role_id = ar.id
        WHERE arm.employee_position_id = $1
          AND arm.context IN (%s)`, strings.Join(placeholders, ","))

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *ActorRoleMappingRepository) GetRoleIDsForUserContext(positionID int, contexts []string) ([]int, error) {
	if len(contexts) == 0 {
		return []int{}, nil
	}

	placeholders := make([]string, len(contexts))
	args := make([]interface{}, len(contexts)+1)
	args[0] = positionID
	for i, context := range contexts {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = context
	}

	query := fmt.Sprintf(`
        SELECT actor_role_id
        FROM actor_role_mapping
        WHERE employee_position_id = $1
          AND context IN (%s)`, strings.Join(placeholders, ","))

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roleIDs []int
	for rows.Next() {
		var roleID int
		if err := rows.Scan(&roleID); err != nil {
			return nil, err
		}
		roleIDs = append(roleIDs, roleID)
	}
	
	return roleIDs, nil
}