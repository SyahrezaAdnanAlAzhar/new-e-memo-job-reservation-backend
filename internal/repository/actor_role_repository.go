package repository

import (
	"database/sql"

	"github.com/lib/pq"
)

type ActorRoleRepository struct {
	DB *sql.DB
}

func NewActorRoleRepository(db *sql.DB) *ActorRoleRepository {
	return &ActorRoleRepository{DB: db}
}

func (r *ActorRoleRepository) GetRoleNameByID(id int) (string, error) {
	var name string
	query := "SELECT name FROM actor_role WHERE id = $1"
	err := r.DB.QueryRow(query, id).Scan(&name)
	return name, err
}

func (r *ActorRoleRepository) GetRoleIDsByNames(names []string) ([]int, error) {
	if len(names) == 0 {
		return []int{}, nil
	}

	query := "SELECT id FROM actor_role WHERE name = ANY($1)"

	rows, err := r.DB.Query(query, pq.Array(names))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}
