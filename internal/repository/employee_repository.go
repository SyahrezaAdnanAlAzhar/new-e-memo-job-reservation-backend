package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
)

type EmployeeRepository struct {
	DB *sql.DB
}

func NewEmployeeRepository(db *sql.DB) *EmployeeRepository {
	return &EmployeeRepository{DB: db}
}

func (r *EmployeeRepository) FindAll(filters dto.EmployeeFilter) ([]dto.EmployeeDetailResponse, int64, error) {
	baseQuery := `
        FROM employee e
        LEFT JOIN department d ON e.department_id = d.id
        LEFT JOIN area a ON e.area_id = a.id
        LEFT JOIN employee_position p ON e.employee_position_id = p.id
    `
	selectQuery := `
        SELECT e.npk, e.name, e.is_active, e.department_id, d.name as department_name,
               e.area_id, a.name as area_name, e.employee_position_id, p.name as position_name
    ` + baseQuery

	countQuery := "SELECT COUNT(e.npk)" + baseQuery

	var conditions []string
	var args []interface{}
	argID := 1

	if filters.DepartmentID != 0 {
		conditions = append(conditions, fmt.Sprintf("e.department_id = $%d", argID))
		args = append(args, filters.DepartmentID)
		argID++
	}
	if filters.AreaID != 0 {
		conditions = append(conditions, fmt.Sprintf("e.area_id = $%d", argID))
		args = append(args, filters.AreaID)
		argID++
	}
	if filters.EmployeePositionID != 0 {
		conditions = append(conditions, fmt.Sprintf("e.employee_position_id = $%d", argID))
		args = append(args, filters.EmployeePositionID)
		argID++
	}
	if filters.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("e.is_active = $%d", argID))
		args = append(args, *filters.IsActive)
		argID++
	}

	if filters.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(e.name ILIKE $%d OR e.npk ILIKE $%d)", argID, argID+1))
		args = append(args, "%"+filters.Search+"%", "%"+filters.Search+"%")
		argID += 2
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	var totalItems int64
	err := r.DB.QueryRow(countQuery+whereClause, args...).Scan(&totalItems)
	if err != nil {
		return nil, 0, err
	}

	if totalItems == 0 {
		return []dto.EmployeeDetailResponse{}, 0, nil
	}

	query := selectQuery + whereClause + " ORDER BY e.npk ASC"

	limitArg := argID
	offsetArg := argID + 1
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", limitArg, offsetArg)
	args = append(args, filters.Limit, (filters.Page-1)*filters.Limit)

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var employees []dto.EmployeeDetailResponse
	for rows.Next() {
		var e dto.EmployeeDetailResponse
		var areaName sql.NullString
		err := rows.Scan(
			&e.NPK, &e.Name, &e.IsActive, &e.DepartmentID, &e.DepartmentName,
			&e.AreaID, &areaName, &e.Position.ID, &e.Position.Name,
		)
		if err != nil {
			return nil, 0, err
		}
		if areaName.Valid {
			e.AreaName = &areaName.String
		}
		employees = append(employees, e)
	}

	return employees, totalItems, nil
}

func (r *EmployeeRepository) FindByNPK(npk string) (*model.Employee, error) {
	query := `
        SELECT 
            e.npk, e.department_id, e.area_id, e.name, e.is_active,
            ep.id as position_id, ep.name as position_name
        FROM employee e
        JOIN employee_position ep ON e.employee_position_id = ep.id
        WHERE e.npk = $1`
	row := r.DB.QueryRow(query, npk)

	var e model.Employee
	err := row.Scan(
		&e.NPK,
		&e.DepartmentID,
		&e.AreaID,
		&e.Name,
		&e.IsActive,
		&e.Position.ID,
		&e.Position.Name,
	)
	if err != nil {
		return nil, err
	}
	if !e.IsActive {
		return nil, errors.New("user is not active")
	}
	return &e, nil
}

func (r *EmployeeRepository) GetEmployeePositionID(ctx context.Context, npk string) (int, error) {
	var positionID int
	query := "SELECT employee_position_id FROM employee WHERE npk = $1"
	err := r.DB.QueryRowContext(ctx, query, npk).Scan(&positionID)
	return positionID, err
}

func (r *EmployeeRepository) FindOptions(filters dto.EmployeeOptionsFilter) ([]dto.EmployeeOptionResponse, error) {
	var query string
	var args []interface{}

	baseQuery := `
        SELECT DISTINCT e.npk, e.name
        FROM employee e
        JOIN ticket t ON %s = e.npk
        JOIN (
            SELECT DISTINCT ON (ticket_id) ticket_id, status_ticket_id
            FROM track_status_ticket
            ORDER BY ticket_id, start_date DESC, id DESC
        ) current_tst ON t.id = current_tst.ticket_id
        JOIN status_ticket st ON current_tst.status_ticket_id = st.id
        WHERE ($1 = 0 OR st.section_id = $1)
          AND ($2 = 0 OR t.department_target_id = $2)
        ORDER BY e.name ASC`

	switch filters.Role {
	case "requestor":
		query = fmt.Sprintf(baseQuery, "t.requestor")
	case "pic":
		baseQuery = `
            SELECT DISTINCT e.npk, e.name
            FROM employee e
            JOIN job j ON j.pic_job = e.npk
            JOIN ticket t ON j.ticket_id = t.id
            JOIN (
                SELECT DISTINCT ON (ticket_id) ticket_id, status_ticket_id
                FROM track_status_ticket
                ORDER BY ticket_id, start_date DESC, id DESC
            ) current_tst ON t.id = current_tst.ticket_id
            JOIN status_ticket st ON current_tst.status_ticket_id = st.id
            WHERE ($1 = 0 OR st.section_id = $1)
              AND ($2 = 0 OR t.department_target_id = $2)
            ORDER BY e.name ASC`
		query = baseQuery
	default:
		return []dto.EmployeeOptionResponse{}, nil
	}

	args = append(args, filters.SectionID, filters.DepartmentTargetID)

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []dto.EmployeeOptionResponse
	for rows.Next() {
		var e dto.EmployeeOptionResponse
		if err := rows.Scan(&e.NPK, &e.Name); err != nil {
			return nil, err
		}
		employees = append(employees, e)
	}
	return employees, nil
}

func (r *EmployeeRepository) Create(req dto.CreateEmployeeRequest) (*model.Employee, error) {
	query := `
        INSERT INTO employee (npk, name, department_id, area_id, employee_position_id, is_active)
        VALUES ($1, $2, $3, $4, $5, true)
        RETURNING npk, name, department_id, area_id, employee_position_id, is_active, created_at, updated_at`

	var areaID sql.NullInt64
	if req.AreaID != nil {
		areaID = sql.NullInt64{Int64: int64(*req.AreaID), Valid: true}
	}

	row := r.DB.QueryRow(query, req.NPK, req.Name, req.DepartmentID, areaID, req.EmployeePositionID)

	var newEmployee model.Employee
	err := row.Scan(
		&newEmployee.NPK, &newEmployee.Name, &newEmployee.DepartmentID, &newEmployee.AreaID,
		&newEmployee.Position.ID, &newEmployee.IsActive, &newEmployee.CreatedAt, &newEmployee.UpdatedAt,
	)
	return &newEmployee, err
}

func (r *EmployeeRepository) Update(npk string, req dto.UpdateEmployeeRequest) (*model.Employee, error) {
	query := `
        UPDATE employee SET name = $1, department_id = $2, area_id = $3, employee_position_id = $4, updated_at = NOW()
        WHERE npk = $5
        RETURNING npk, name, department_id, area_id, employee_position_id, is_active, created_at, updated_at`

	var areaID sql.NullInt64
	if req.AreaID != nil {
		areaID = sql.NullInt64{Int64: int64(*req.AreaID), Valid: true}
	}

	row := r.DB.QueryRow(query, req.Name, req.DepartmentID, areaID, req.EmployeePositionID, npk)

	var updatedEmployee model.Employee
	err := row.Scan(
		&updatedEmployee.NPK, &updatedEmployee.Name, &updatedEmployee.DepartmentID, &updatedEmployee.AreaID,
		&updatedEmployee.Position.ID, &updatedEmployee.IsActive, &updatedEmployee.CreatedAt, &updatedEmployee.UpdatedAt,
	)
	return &updatedEmployee, err
}

func (r *EmployeeRepository) UpdateActiveStatus(npk string, isActive bool) error {
	query := "UPDATE employee SET is_active = $1, updated_at = NOW() WHERE npk = $2"
	result, err := r.DB.Exec(query, isActive, npk)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
