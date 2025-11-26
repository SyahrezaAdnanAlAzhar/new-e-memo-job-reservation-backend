package repository

import (
	"database/sql"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
)

type AppUserRepository struct {
	DB *sql.DB
}

func NewAppUserRepository(db *sql.DB) *AppUserRepository {
	return &AppUserRepository{DB: db}
}

// GET BY USERNAME OR NPK
func (r *AppUserRepository) FindByUsernameOrNPK(username string) (*model.AppUser, error) {
	query := `
        SELECT id, username, password, user_type, employee_npk, employee_position_id
        FROM app_user 
        WHERE (username = $1 OR employee_npk = $1) AND is_active = true`

	row := r.DB.QueryRow(query, username)

	var user model.AppUser
	err := row.Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.UserType,
		&user.EmployeeNPK, &user.EmployeePositionID,
	)
	return &user, err
}

// GET BY ID
func (r *AppUserRepository) GetUserDetailByID(userID int) (*dto.UserDetail, error) {
	query := `
        SELECT
            u.id, u.username, u.user_type, u.employee_npk,
            e.name as employee_name,
            ep.name as employee_position,
            d.name as employee_department,
            a.name as employee_area
        FROM app_user u
        LEFT JOIN employee e ON u.employee_npk = e.npk
        LEFT JOIN employee_position ep ON u.employee_position_id = ep.id
        LEFT JOIN department d ON e.department_id = d.id
        LEFT JOIN area a ON e.area_id = a.id
        WHERE u.id = $1`

	row := r.DB.QueryRow(query, userID)

	var userDetail dto.UserDetail
	var empName, empPos, empDept, empArea sql.NullString

	err := row.Scan(
		&userDetail.UserID, &userDetail.Username, &userDetail.UserType, &userDetail.EmployeeNPK,
		&empName, &empPos, &empDept, &empArea,
	)
	if err != nil {
		return nil, err
	}

	if empName.Valid {
		userDetail.EmployeeName = &empName.String
	}
	if empPos.Valid {
		userDetail.EmployeePosition = &empPos.String
	}
	if empDept.Valid {
		userDetail.EmployeeDepartment = &empDept.String
	}
	if empArea.Valid {
		userDetail.EmployeeArea = &empArea.String
	}

	return &userDetail, nil
}

// GET BY ID
func (r *AppUserRepository) FindByID(id int) (*model.AppUser, error) {
	query := `
        SELECT id, username, password, user_type, employee_npk, employee_position_id
        FROM app_user 
        WHERE id = $1`
	row := r.DB.QueryRow(query, id)
	var user model.AppUser
	err := row.Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.UserType,
		&user.EmployeeNPK, &user.EmployeePositionID,
	)
	return &user, err
}
