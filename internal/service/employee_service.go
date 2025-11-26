package service

import (
	"database/sql"
	"errors"

	"e-memo-job-reservation-api/internal/dto"
	"e-memo-job-reservation-api/internal/model"
	"e-memo-job-reservation-api/internal/repository"

	"github.com/jackc/pgx/v5/pgconn"
)

type EmployeeService struct {
	repo *repository.EmployeeRepository
}

func NewEmployeeService(repo *repository.EmployeeRepository) *EmployeeService {
	return &EmployeeService{repo: repo}
}

func (s *EmployeeService) CreateEmployee(req dto.CreateEmployeeRequest) (*model.Employee, error) {
	newEmployee, err := s.repo.Create(req)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // unique_violation (untuk NPK)
				return nil, errors.New("employee with this NPK already exists")
			}
			if pgErr.Code == "23503" { // foreign_key_violation
				return nil, errors.New("invalid department_id, area_id, or employee_position_id")
			}
		}
		return nil, err
	}
	return newEmployee, nil
}

func (s *EmployeeService) UpdateEmployee(npk string, req dto.UpdateEmployeeRequest) (*model.Employee, error) {
	updatedEmployee, err := s.repo.Update(npk, req)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("employee not found")
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, errors.New("invalid department_id, area_id, or employee_position_id")
		}
		return nil, err
	}
	return updatedEmployee, nil
}

func (s *EmployeeService) UpdateEmployeeActiveStatus(npk string, req dto.UpdateEmployeeStatusRequest) error {
	err := s.repo.UpdateActiveStatus(npk, req.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("employee not found")
		}
		return err
	}
	return nil
}

func (s *EmployeeService) GetAllEmployees(filters dto.EmployeeFilter) (*dto.PaginatedEmployeeResponse, error) {
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.Limit <= 0 {
		filters.Limit = 10
	}

	employees, totalItems, err := s.repo.FindAll(filters)
	if err != nil {
		return nil, err
	}

	totalPages := 0
	if totalItems > 0 {
		totalPages = int((totalItems + int64(filters.Limit) - 1) / int64(filters.Limit))
	}

	paginatedResponse := &dto.PaginatedEmployeeResponse{
		Data: employees,
		Pagination: dto.Pagination{
			CurrentPage: filters.Page,
			TotalPages:  totalPages,
			TotalItems:  totalItems,
			PageSize:    filters.Limit,
		},
	}

	return paginatedResponse, nil
}

func (s *EmployeeService) GetEmployeeOptions(filters dto.EmployeeOptionsFilter) ([]dto.EmployeeOptionResponse, error) {
	return s.repo.FindOptions(filters)
}
