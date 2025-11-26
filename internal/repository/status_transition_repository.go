package repository

import (
	"database/sql"
	"fmt"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"

	"github.com/lib/pq"
)

type StatusTransitionRepository struct {
	DB *sql.DB
}

func NewStatusTransitionRepository(db *sql.DB) *StatusTransitionRepository {
	return &StatusTransitionRepository{DB: db}
}

type TransitionDetail struct {
	RequiredActorRole string
	ActionDetail      dto.AvailableTicketActionResponse
}

func (r *StatusTransitionRepository) FindValidTransition(fromStatusID int, actionName string) (int, []int, error) {
	query := `
        SELECT 
            st.to_status_id,
            st.actor_role_id
        FROM status_transition st
        JOIN action a ON st.action_id = a.id
        WHERE st.from_status_id = $1 AND a.name = $2 AND st.is_active = true`

	rows, err := r.DB.Query(query, fromStatusID, actionName)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	var toStatusID int
	var allowedRoleIDs []int
	isToStatusSet := false

	for rows.Next() {
		var currentToStatusID, currentActorRoleID int
		if err := rows.Scan(&currentToStatusID, &currentActorRoleID); err != nil {
			return 0, nil, err
		}

		if !isToStatusSet {
			toStatusID = currentToStatusID
			isToStatusSet = true
		} else if toStatusID != currentToStatusID {
			return 0, nil, fmt.Errorf("data inconsistency: action '%s' from status %d leads to multiple different to_statuses", actionName, fromStatusID)
		}

		allowedRoleIDs = append(allowedRoleIDs, currentActorRoleID)
	}

	if !isToStatusSet {
		return 0, nil, sql.ErrNoRows
	}

	return toStatusID, allowedRoleIDs, nil
}

func (r *StatusTransitionRepository) FindPossibleTransitionsWithDetails(fromStatusID int) ([]TransitionDetail, error) {
	query := `
        SELECT 
            ar.name as required_actor_role,
            a.name as action_name,
            a.hex_code,
            st.require_reason,
            st.reason_label,
            st.require_file
        FROM status_transition st
        JOIN action a ON st.action_id = a.id
        JOIN actor_role ar ON st.actor_role_id = ar.id
        WHERE st.from_status_id = $1 AND st.is_active = true AND a.is_active = true`

	rows, err := r.DB.Query(query, fromStatusID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transitions []TransitionDetail
	for rows.Next() {
		var t TransitionDetail
		err := rows.Scan(
			&t.RequiredActorRole,
			&t.ActionDetail.ActionName,
			&t.ActionDetail.HexCode,
			&t.ActionDetail.RequireReason,
			&t.ActionDetail.ReasonLabel,
			&t.ActionDetail.RequireFile,
		)
		if err != nil {
			return nil, err
		}
		transitions = append(transitions, t)
	}
	return transitions, nil
}

func (r *StatusTransitionRepository) FindAvailableTransitionsForRoles(fromStatusID int, roleIDs []int) ([]dto.AvailableTicketActionResponse, error) {
	if len(roleIDs) == 0 {
		return []dto.AvailableTicketActionResponse{}, nil
	}

	query := `
        SELECT 
            a.name as action_name,
            a.id as action_id,
            st.to_status_id,
            a.hex_code,
            st.require_reason,
            st.reason_label,
            st.require_file
        FROM status_transition st
        JOIN action a ON st.action_id = a.id
        WHERE st.from_status_id = $1
          AND st.actor_role_id = ANY($2)
          AND st.is_active = true
          AND a.is_active = true`

	rows, err := r.DB.Query(query, fromStatusID, pq.Array(roleIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []dto.AvailableTicketActionResponse
	for rows.Next() {
		var a dto.AvailableTicketActionResponse
		if err := rows.Scan(
			&a.ActionName,
			&a.ActionID,
			&a.ToStatusID,
			&a.HexCode,
			&a.RequireReason,
			&a.ReasonLabel,
			&a.RequireFile,
		); err != nil {
			return nil, err
		}
		actions = append(actions, a)
	}
	return actions, nil
}

func (r *StatusTransitionRepository) GetTransitionDetails(fromStatusID int, actionName string) (*model.StatusTransition, error) {
	query := `
        SELECT 
            st.id, st.from_status_id, st.to_status_id, st.action_id, 
            st.actor_role_id, st.require_reason, st.reason_label, st.require_file
        FROM status_transition st
        JOIN action a ON st.action_id = a.id
        WHERE st.from_status_id = $1 AND a.name = $2 AND st.is_active = true
        LIMIT 1`

	row := r.DB.QueryRow(query, fromStatusID, actionName)

	var transition model.StatusTransition
	err := row.Scan(
		&transition.ID, &transition.FromStatusID, &transition.ToStatusID, &transition.ActionID,
		&transition.ActorRoleID, &transition.RequiresReason, &transition.ReasonLabel, &transition.RequiresFile,
	)
	if err != nil {
		return nil, err
	}
	return &transition, nil
}

func (r *StatusTransitionRepository) HasAvailableActionsForRoles(fromStatusID int, actorRoleIDs []int) (bool, error) {
	if len(actorRoleIDs) == 0 {
		return false, nil
	}

	var exists bool
	query := `
        SELECT EXISTS (
            SELECT 1 FROM status_transition
            WHERE from_status_id = $1
              AND actor_role_id = ANY($2)
              AND is_active = true
        )`

	err := r.DB.QueryRow(query, fromStatusID, pq.Array(actorRoleIDs)).Scan(&exists)
	return exists, err
}
