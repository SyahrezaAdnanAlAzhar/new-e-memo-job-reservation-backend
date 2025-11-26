package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"e-memo-job-reservation-api/internal/model"

	"github.com/lib/pq"
)

type JobRepository struct {
	DB *sql.DB
}

func NewJobRepository(db *sql.DB) *JobRepository {
	return &JobRepository{DB: db}
}

// CREATE
func (r *JobRepository) Create(ctx context.Context, tx *sql.Tx, ticketID int, initialJobPriority int) error {
	query := "INSERT INTO job (ticket_id, job_priority) VALUES ($1, $2)"
	_, err := tx.ExecContext(ctx, query, ticketID, initialJobPriority)
	return err
}

// CHECK THE JOB ALREADY GET ASSIGN OR NOT
func (r *JobRepository) IsJobAssigned(ctx context.Context, ticketID int) (bool, error) {
	var isAssigned bool
	query := "SELECT (pic_job IS NOT NULL) FROM job WHERE ticket_id = $1"

	err := r.DB.QueryRowContext(ctx, query, ticketID).Scan(&isAssigned)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return isAssigned, nil
}

// GET PIC
func (r *JobRepository) GetPicByTicketID(ctx context.Context, ticketID int) (string, error) {
	var picNpk sql.NullString
	query := "SELECT pic_job FROM job WHERE ticket_id = $1"

	err := r.DB.QueryRowContext(ctx, query, ticketID).Scan(&picNpk)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("job not found for this ticket")
		}
		return "", err
	}

	if !picNpk.Valid {
		return "", nil
	}

	return picNpk.String, nil
}

// AssignPIC
func (r *JobRepository) AssignPIC(id int, picNpk string) error {
	query := "UPDATE job SET pic_job = $1, updated_at = NOW() WHERE id = $2"
	result, err := r.DB.Exec(query, picNpk, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// UpdatePriority
func (r *JobRepository) UpdatePriority(ctx context.Context, tx *sql.Tx, jobID int, version int, newPriority int) (int64, error) {
	query := `
        UPDATE job 
        SET job_priority = $1, version = version + 1, updated_at = NOW()
        WHERE id = $2 AND version = $3`

	result, err := tx.ExecContext(ctx, query, newPriority, jobID, version)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// CheckJobsInDepartment
func (r *JobRepository) CheckJobsInDepartment(jobIDs []int, departmentID int) (int, error) {
	if len(jobIDs) == 0 {
		return 0, nil
	}

	query := `
        SELECT COUNT(j.id) 
        FROM job j
        JOIN ticket t ON j.ticket_id = t.id
        WHERE j.id = ANY($1) AND t.department_target_id = $2`

	var count int
	err := r.DB.QueryRow(query, pq.Array(jobIDs), departmentID).Scan(&count)
	return count, err
}

func (r *JobRepository) ForceUpdatePriority(ctx context.Context, tx *sql.Tx, jobID int, newPriority int) error {
	query := `
        UPDATE job 
        SET job_priority = $1, version = version + 1, updated_at = NOW()
        WHERE id = $2`

	_, err := tx.ExecContext(ctx, query, newPriority, jobID)
	return err
}

func (r *JobRepository) AddReportFile(id int, filePath string) error {
	query := "UPDATE job SET report_file = array_append(report_file, $1), updated_at = NOW() WHERE id = $2"

	result, err := r.DB.Exec(query, filePath, id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GET JOB BY TICKET ID
func (r *JobRepository) FindByTicketID(ctx context.Context, ticketID int) (*model.Job, error) {
	query := `
        SELECT id, ticket_id, pic_job, job_priority, report_file, version, created_at, updated_at 
        FROM job
        WHERE ticket_id = $1`

	row := r.DB.QueryRowContext(ctx, query, ticketID)

	var j model.Job
	var reportFilesJSON []byte

	err := row.Scan(
		&j.ID,
		&j.TicketID,
		&j.PicJob,
		&j.JobPriority,
		&reportFilesJSON,
		&j.Version,
		&j.CreatedAt,
		&j.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(reportFilesJSON) > 0 {
		if err := json.Unmarshal(reportFilesJSON, &j.ReportFiles); err != nil {
			log.Printf("WARNING: Failed to unmarshal report_file for job with ticket_id %d: %v", ticketID, err)
		}
	}

	return &j, nil
}

func (r *JobRepository) FindByID(id int) (*model.Job, error) {
	query := `
        SELECT id, ticket_id, pic_job, job_priority, report_file, version, created_at, updated_at 
        FROM job WHERE id = $1`

	row := r.DB.QueryRow(query, id)

	var j model.Job
	err := row.Scan(
		&j.ID, &j.TicketID, &j.PicJob, &j.JobPriority, &j.ReportFiles, &j.Version, &j.CreatedAt, &j.UpdatedAt,
	)
	return &j, err
}

// [GANTI FUNGSI INI]
func (r *JobRepository) UpdateJobCompletionDetails(ctx context.Context, tx *sql.Tx, ticketID int, filesMetadata []model.FileMetadata, spendingAmount *int64) error {
	if len(filesMetadata) > 0 {
		// Convert all file metadata to JSONB array
		var fileMetadataJSON []interface{}
		for _, fm := range filesMetadata {
			fileMetadataJSON = append(fileMetadataJSON, fm)
		}

		jsonBytes, err := json.Marshal(fileMetadataJSON)
		if err != nil {
			return fmt.Errorf("failed to marshal file metadata: %w", err)
		}

		// Replace the entire report_file array instead of appending
		query := `
            UPDATE job 
            SET report_file = $1::jsonb, 
                updated_at = NOW()
            WHERE ticket_id = $2`

		result, err := tx.ExecContext(ctx, query, string(jsonBytes), ticketID)
		if err != nil {
			return err
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return sql.ErrNoRows
		}
	}

	if spendingAmount != nil {
		query := `
            UPDATE job 
            SET spending_amount = $1, updated_at = NOW()
            WHERE ticket_id = $2`

		result, err := tx.ExecContext(ctx, query, *spendingAmount, ticketID)
		if err != nil {
			return err
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return sql.ErrNoRows
		}
	}

	return nil
}
