package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"

	"github.com/lib/pq"
)

type TicketRepository struct {
	DB *sql.DB
}

func NewTicketRepository(db *sql.DB) *TicketRepository {
	return &TicketRepository{DB: db}
}

const baseTicketQuery = `
    SELECT
        t.id,
        t.description,
        t.department_target_id,
        dt.name as department_target_name,
        t.ticket_priority,
        t.version,
        j.id as job_id,
        j.job_priority,
        pl.name as location_name,
        sl.name as specified_location_name,
        t.created_at,
        (NOW()::date - t.created_at::date) as ticket_age_days,
        t.deadline,
        (t.deadline::date - NOW()::date) as days_remaining,
        req_emp.name as requestor_name,
		req_emp.npk as requestor_npk,
        req_dept.name as requestor_department,
        pic_emp.name as pic_name,
		pic_emp.npk as pic_npk,
        pic_area.name as pic_area_name,
        current_st.name as current_status,
        current_st.hex_color as current_status_hex_code,
        current_sst.name as current_section_name
    FROM ticket t
    LEFT JOIN job j ON t.id = j.ticket_id
    LEFT JOIN department dt ON t.department_target_id = dt.id
    LEFT JOIN physical_location pl ON t.physical_location_id = pl.id
    LEFT JOIN specified_location sl ON t.specified_location_id = sl.id
    JOIN employee req_emp ON t.requestor = req_emp.npk
    LEFT JOIN department req_dept ON req_emp.department_id = req_dept.id
    LEFT JOIN employee pic_emp ON j.pic_job = pic_emp.npk
    LEFT JOIN area pic_area ON pic_emp.area_id = pic_area.id
    LEFT JOIN (
        SELECT DISTINCT ON (ticket_id) ticket_id, status_ticket_id
        FROM track_status_ticket
        ORDER BY ticket_id, start_date DESC
    ) current_tst ON t.id = current_tst.ticket_id
    LEFT JOIN status_ticket current_st ON current_tst.status_ticket_id = current_st.id
    LEFT JOIN section_status_ticket current_sst ON current_st.section_id = current_sst.id
`

// MAIN

// CREATE TICKET
func (r *TicketRepository) Create(ctx context.Context, tx *sql.Tx, ticket model.Ticket) (*model.Ticket, error) {
	supportFilesJSON, _ := json.Marshal(ticket.SupportFiles)

	query := `
        INSERT INTO ticket (
            requestor, department_target_id, physical_location_id, 
            specified_location_id, description, ticket_priority, deadline, support_file
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, created_at, updated_at`

	row := tx.QueryRowContext(ctx, query,
		ticket.Requestor,
		ticket.DepartmentTargetID,
		ticket.PhysicalLocationID,
		ticket.SpecifiedLocationID,
		ticket.Description,
		ticket.TicketPriority,
		ticket.Deadline,
		supportFilesJSON,
	)

	var newTicket model.Ticket = ticket
	err := row.Scan(&newTicket.ID, &newTicket.CreatedAt, &newTicket.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &newTicket, nil
}

// GET ALL
func (r *TicketRepository) FindAll(filters dto.TicketFilter) ([]dto.TicketDetailResponse, error) {
	query := baseTicketQuery

	var conditions []string
	var args []interface{}
	argID := 1

	if filters.SectionID != 0 {
		conditions = append(conditions, fmt.Sprintf("current_sst.id = $%d", argID))
		args = append(args, filters.SectionID)
		argID++
	}

	if len(filters.StatusID) > 0 {
		conditions = append(conditions, fmt.Sprintf("current_st.id = ANY($%d)", argID))
		args = append(args, pq.Array(filters.StatusID))
		argID++
	}

	if filters.DepartmentTargetID != 0 {
		conditions = append(conditions, fmt.Sprintf("t.department_target_id = $%d", argID))
		args = append(args, filters.DepartmentTargetID)
		argID++
	}

	if len(filters.RequestorDepartmentID) > 0 {
		conditions = append(conditions, fmt.Sprintf("req_emp.department_id = ANY($%d)", argID))
		args = append(args, pq.Array(filters.RequestorDepartmentID))
		argID++
	}

	if filters.DepartmentTargetName != "" {
		conditions = append(conditions, fmt.Sprintf("dt.name ILIKE $%d", argID))
		args = append(args, "%"+filters.DepartmentTargetName+"%")
		argID++
	}

	if len(filters.Requestor) > 0 {
		conditions = append(conditions, fmt.Sprintf("t.requestor = ANY($%d)", argID))
		args = append(args, pq.Array(filters.Requestor))
		argID++
	}

	if len(filters.PicNPK) > 0 {
		conditions = append(conditions, fmt.Sprintf("j.pic_job = ANY($%d)", argID))
		args = append(args, pq.Array(filters.PicNPK))
		argID++
	}

	if filters.Year != 0 {
		conditions = append(conditions, fmt.Sprintf("EXTRACT(YEAR FROM t.created_at) = $%d", argID))
		args = append(args, filters.Year)
		argID++
	}
	if filters.Month != 0 {
		conditions = append(conditions, fmt.Sprintf("EXTRACT(MONTH FROM t.created_at) = $%d", argID))
		args = append(args, filters.Month)
		argID++
	}

	if filters.SearchQuery != "" {
		searchQuery := strings.ReplaceAll(strings.TrimSpace(filters.SearchQuery), " ", " | ")

		conditions = append(conditions, fmt.Sprintf("t.description_tsv @@ to_tsquery('simple', $%d)", argID))
		args = append(args, searchQuery)
		argID++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	orderByClause := " ORDER BY t.ticket_priority ASC"
	if filters.SortBy != "" {
		allowedSortColumns := map[string]string{
			"priority":  "t.ticket_priority",
			"deadline":  "t.deadline",
			"age":       "ticket_age_days",
			"status":    "current_st.name",
			"requestor": "req_emp.name",
			"pic":       "pic_emp.name",
		}

		var sortClauses []string

		sortParams := strings.Split(filters.SortBy, ",")

		for _, param := range sortParams {
			parts := strings.Split(strings.TrimSpace(param), "_")
			if len(parts) != 2 {
				continue
			}

			columnKey := parts[0]
			direction := strings.ToUpper(parts[1])

			if dbColumn, ok := allowedSortColumns[columnKey]; ok && (direction == "ASC" || direction == "DESC") {
				sortClauses = append(sortClauses, dbColumn+" "+direction)
			}
		}

		if len(sortClauses) > 0 {
			orderByClause = " ORDER BY " + strings.Join(sortClauses, ", ")
		}
	}
	query += orderByClause

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTicketDetails(rows)
}

// GET BY ID
func (r *TicketRepository) FindByID(id int) (*dto.TicketDetailResponse, error) {
	query := baseTicketQuery + " WHERE t.id = $1"
	rows, err := r.DB.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tickets, err := scanTicketDetails(rows)
	if err != nil {
		return nil, err
	}
	if len(tickets) == 0 {
		return nil, sql.ErrNoRows
	}
	return &tickets[0], nil
}

// GET BY ID AS STRUCT
func (r *TicketRepository) FindByIDAsStruct(ctx context.Context, id int) (*model.Ticket, error) {
	query := "SELECT id, requestor, department_target_id, physical_location_id, specified_location_id, description, ticket_priority FROM ticket WHERE id = $1"
	row := r.DB.QueryRowContext(ctx, query, id)

	var t model.Ticket
	err := row.Scan(&t.ID, &t.Requestor, &t.DepartmentTargetID, &t.PhysicalLocationID, &t.SpecifiedLocationID, &t.Description, &t.TicketPriority)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// UPDATE TICKET
func (r *TicketRepository) Update(ctx context.Context, tx *sql.Tx, id int, req dto.UpdateTicketRequest, specifiedLocationID sql.NullInt64) (int64, error) {
	query := `
        UPDATE ticket 
        SET 
            department_target_id = $1, 
            description = $2, 
            physical_location_id = $3, 
            specified_location_id = $4, 
            deadline = $5, 
            version = version + 1,
            updated_at = NOW()
        WHERE id = $6 AND version = $7`

	deadline, err := ParseDeadline(req.Deadline)
	if err != nil {
		return 0, err
	}

	result, err := tx.ExecContext(ctx, query,
		req.DepartmentTargetID,
		req.Description,
		toNullInt64(req.PhysicalLocationID),
		specifiedLocationID,
		deadline,
		id,
		req.Version,
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// REORDER
func (r *TicketRepository) UpdatePriority(ctx context.Context, tx *sql.Tx, ticketID int, version int, newPriority int) (int64, error) {
	query := `
        UPDATE ticket 
        SET ticket_priority = $1, version = version + 1, updated_at = NOW()
        WHERE id = $2 AND version = $3`

	result, err := tx.ExecContext(ctx, query, newPriority, ticketID, version)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// FORCE REORDER
func (r *TicketRepository) ForceUpdatePriority(ctx context.Context, tx *sql.Tx, ticketID int, newPriority int) error {
	query := `
        UPDATE ticket 
        SET ticket_priority = $1, version = version + 1, updated_at = NOW()
        WHERE id = $2`

	_, err := tx.ExecContext(ctx, query, newPriority, ticketID)
	return err
}

// UPDATE TICKET TO FALLBACK STATUS
func (r *TicketRepository) MoveTicketsToFallbackStatus(ctx context.Context, tx *sql.Tx, sectionIDToDeactivate int, fallbackStatusID int) error {
	findTicketsQuery := `
        SELECT tst.ticket_id
        FROM track_status_ticket tst
        WHERE tst.finish_date IS NULL
        AND tst.status_ticket_id IN (
            SELECT id FROM status_ticket WHERE section_id = $1
        )`
	rows, err := tx.QueryContext(ctx, findTicketsQuery, sectionIDToDeactivate)
	if err != nil {
		return err
	}
	defer rows.Close()

	var ticketIDsToMove []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return err
		}
		ticketIDsToMove = append(ticketIDsToMove, id)
	}

	if len(ticketIDsToMove) == 0 {
		return nil
	}

	deleteQuery := `
        DELETE FROM track_status_ticket
        WHERE ticket_id = ANY($1)
        AND status_ticket_id IN (
            SELECT id FROM status_ticket WHERE section_id = $2
        )`
	_, err = tx.ExecContext(ctx, deleteQuery, ticketIDsToMove, sectionIDToDeactivate)
	if err != nil {
		return err
	}

	createQuery := "INSERT INTO track_status_ticket (ticket_id, status_ticket_id, start_date) VALUES ($1, $2, NOW())"
	stmt, err := tx.PrepareContext(ctx, createQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, ticketID := range ticketIDsToMove {
		_, err := stmt.ExecContext(ctx, ticketID, fallbackStatusID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *TicketRepository) CheckTicketsFromDepartment(ticketIDs []int, requestorDepartmentID int) (int, error) {
	if len(ticketIDs) == 0 {
		return 0, nil
	}

	query := `
        SELECT COUNT(t.id) 
        FROM ticket t
        JOIN employee e ON t.requestor = e.npk
        WHERE t.id = ANY($1) AND e.department_id = $2`

	var count int
	err := r.DB.QueryRow(query, pq.Array(ticketIDs), requestorDepartmentID).Scan(&count)
	return count, err
}

// HELPER

// GET LAST PRIORITY
func (r *TicketRepository) GetLastPriority(ctx context.Context, tx *sql.Tx, departmentTargetID int) (int, error) {
	var lastPriority sql.NullInt64
	query := "SELECT MAX(ticket_priority) FROM ticket WHERE department_target_id = $1"
	err := tx.QueryRowContext(ctx, query, departmentTargetID).Scan(&lastPriority)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if !lastPriority.Valid {
		return 1, nil
	}
	return int(lastPriority.Int64) + 1, nil
}

func toNullInt64(val *int) sql.NullInt64 {
	if val == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: int64(*val), Valid: true}
}

// PARSE DEADLINE
func ParseDeadline(deadlineStr *string) (sql.NullTime, error) {
	if deadlineStr == nil {
		return sql.NullTime{Valid: false}, nil
	}
	// Format "2006-01-02"
	t, err := time.Parse("2006-01-02", *deadlineStr)
	if err != nil {
		return sql.NullTime{Valid: false}, err
	}
	return sql.NullTime{Time: t, Valid: true}, nil
}

// QUERY MAPPING
func scanTicketDetails(rows *sql.Rows) ([]dto.TicketDetailResponse, error) {
	var tickets []dto.TicketDetailResponse
	for rows.Next() {
		var t dto.TicketDetailResponse
		err := rows.Scan(
			&t.TicketID,
			&t.Description,
			&t.DepartmentTargetID,
			&t.DepartmentTargetName,
			&t.TicketPriority,
			&t.Version,
			&t.JobID,
			&t.JobPriority,
			&t.LocationName,
			&t.SpecifiedLocationName,
			&t.CreatedAt,
			&t.TicketAgeDays,
			&t.Deadline,
			&t.DaysRemaining,
			&t.RequestorName,
			&t.RequestorNPK,
			&t.RequestorDepartment,
			&t.PicName,
			&t.PicNPK,
			&t.PicAreaName,
			&t.CurrentStatus,
			&t.CurrentStatusHexCode,
			&t.CurrentSectionName,
		)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}
	return tickets, nil
}

// ADD SUPPORT FILE FOR TICKET
func (r *TicketRepository) AddSupportFiles(ctx context.Context, ticketID int, filesMetadata []model.FileMetadata) error {
	if len(filesMetadata) == 0 {
		return nil
	}
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, fm := range filesMetadata {
		jsonBytes, err := json.Marshal(fm)
		if err != nil {
			return fmt.Errorf("failed to marshal file metadata: %w", err)
		}

		query := `
            UPDATE ticket 
            SET support_file = COALESCE(support_file, '[]'::jsonb) || $1::jsonb, 
                updated_at = NOW()
            WHERE id = $2`

		result, err := tx.ExecContext(ctx, query, string(jsonBytes), ticketID)
		if err != nil {
			return err
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return sql.ErrNoRows
		}
	}

	return tx.Commit()
}

func (r *TicketRepository) RemoveSupportFiles(ctx context.Context, ticketID int, filePathsToRemove []string) error {
	if len(filePathsToRemove) == 0 {
		return nil
	}

	query := `
        WITH files_to_remove (path) AS (
            SELECT unnest($1::text[])
        ),
        updated_files AS (
            SELECT jsonb_agg(elem) as new_array
            FROM ticket, jsonb_array_elements(COALESCE(support_file, '[]'::jsonb)) as elem
            WHERE id = $2
            AND (elem ->> 'file_path') NOT IN (SELECT path FROM files_to_remove)
        )
        UPDATE ticket
        SET 
            support_file = (SELECT new_array FROM updated_files),
            updated_at = NOW()
        WHERE id = $2`

	result, err := r.DB.ExecContext(ctx, query, pq.Array(filePathsToRemove), ticketID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *TicketRepository) GetTicketSummary(filters dto.TicketSummaryFilter) ([]dto.TicketSummaryResponse, error) {
	baseQuery := `
        SELECT
            st.id as status_id,
            st.name as status_name,
            st.hex_color,
            COALESCE(ticket_counts.total, 0) as total
        FROM status_ticket st
        LEFT JOIN (
            SELECT
                current_tst.status_ticket_id,
                COUNT(t.id) as total
            FROM ticket t
            JOIN (
                SELECT DISTINCT ON (ticket_id) ticket_id, status_ticket_id
                FROM track_status_ticket
                ORDER BY ticket_id, start_date DESC, id DESC
            ) current_tst ON t.id = current_tst.ticket_id
            %s 
            GROUP BY current_tst.status_ticket_id
        ) ticket_counts ON st.id = ticket_counts.status_ticket_id
    `
	var conditions []string
	var args []interface{}
	argID := 1

	if filters.DepartmentID != 0 {
		conditions = append(conditions, fmt.Sprintf("t.department_target_id = $%d", argID))
		args = append(args, filters.DepartmentID)
		argID++
	}

	year := filters.Year
	month := filters.Month

	if year == 0 {
		year = time.Now().Year()
	}

	conditions = append(conditions, fmt.Sprintf("EXTRACT(YEAR FROM t.created_at) = $%d", argID))
	args = append(args, year)
	argID++

	if month != 0 {
		conditions = append(conditions, fmt.Sprintf("EXTRACT(MONTH FROM t.created_at) = $%d", argID))
		args = append(args, month)
		argID++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(baseQuery, whereClause)

	if filters.SectionID != 0 {
		query += fmt.Sprintf(" WHERE st.section_id = $%d", argID)
		args = append(args, filters.SectionID)
		argID++
	}

	query += " ORDER BY st.sequence ASC"

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summary []dto.TicketSummaryResponse
	for rows.Next() {
		var s dto.TicketSummaryResponse
		if err := rows.Scan(&s.StatusID, &s.StatusName, &s.HexCode, &s.Total); err != nil {
			return nil, err
		}
		summary = append(summary, s)
	}
	return summary, nil
}

func (r *TicketRepository) FindOldestTicket() (*dto.OldestTicketResponse, error) {
	query := "SELECT id, created_at FROM ticket ORDER BY created_at ASC LIMIT 1"
	row := r.DB.QueryRow(query)

	var oldestTicket dto.OldestTicketResponse
	err := row.Scan(&oldestTicket.TicketID, &oldestTicket.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &oldestTicket, nil
}

func (r *TicketRepository) GetSupportFilesByTicketID(ctx context.Context, ticketID int) ([]model.FileMetadata, time.Time, error) {
	var jsonFiles []byte
	var updatedAt time.Time

	query := "SELECT COALESCE(support_file, '[]'::jsonb), updated_at FROM ticket WHERE id = $1"

	err := r.DB.QueryRowContext(ctx, query, ticketID).Scan(&jsonFiles, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, time.Time{}, errors.New("ticket not found")
		}
		return nil, time.Time{}, err
	}

	var filesMetadata []model.FileMetadata
	if string(jsonFiles) == "null" || len(jsonFiles) == 0 || string(jsonFiles) == "[]" {
		return []model.FileMetadata{}, updatedAt, nil
	}
	if err := json.Unmarshal(jsonFiles, &filesMetadata); err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to unmarshal support files: %w", err)
	}

	return filesMetadata, updatedAt, nil
}
