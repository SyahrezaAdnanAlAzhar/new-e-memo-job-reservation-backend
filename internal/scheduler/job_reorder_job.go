package scheduler

import (
	"context"
	"database/sql"
	"log"
	"sort"
	"time"

	"e-memo-job-reservation-api/internal/model"
	"e-memo-job-reservation-api/internal/repository"
	"e-memo-job-reservation-api/internal/websocket"

	"github.com/gin-gonic/gin"
)

type jobWithScore struct {
	ID    int
	Score float64
}

type JobReorderJob struct {
	jobRepo *repository.JobRepository
	db      *sql.DB
	hub     *websocket.Hub
}

func NewJobReorderJob(db *sql.DB, jobRepo *repository.JobRepository, hub *websocket.Hub) *JobReorderJob {
	return &JobReorderJob{db: db, jobRepo: jobRepo, hub: hub}
}

func (j *JobReorderJob) Run() {
	log.Println("Starting JOB priority recalculation job...")
	ctx := context.Background()

	departmentIDs, err := j.getActiveTargetDepartments(ctx)
	if err != nil {
		log.Printf("ERROR (Job Reorder): Could not get target departments: %v", err)
		return
	}
	for _, deptID := range departmentIDs {
		log.Printf("Processing JOBS for department ID: %d", deptID)
		err := j.reorderJobsForDepartment(ctx, deptID)
		if err != nil {
			log.Printf("ERROR (Job Reorder): Failed to reorder jobs for department %d: %v", deptID, err)
			continue
		}
	}

	payload := gin.H{"message": "Job priorities have been recalculated by the system."}
	message, err := websocket.NewMessage("JOB_PRIORITY_RECALCULATED", payload)
	if err != nil {
		log.Printf("CRITICAL: Failed to create websocket message for job cron job: %v", err)
	} else {
		j.hub.BroadcastMessage(message)
	}

	log.Println("JOB priority recalculation job finished.")
}

func (j *JobReorderJob) reorderJobsForDepartment(ctx context.Context, departmentID int) error {
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	jobs, err := j.getActiveJobsByDepartment(ctx, tx, departmentID)
	if err != nil {
		return err
	}
	if len(jobs) == 0 {
		log.Printf("No active jobs for department ID: %d. Skipping.", departmentID)
		return nil
	}

	scoredJobs := make([]jobWithScore, len(jobs))
	for i, job := range jobs {
		ageInDays := time.Since(job.Ticket.CreatedAt).Hours() / 24

		ageWeight := calculateAgeWeight(ageInDays)
		deadlineWeight := calculateDeadlineWeight(job.Ticket.Deadline)

		jobPriorityWeight := 2.0 / float64(job.JobPriority)

		ticketPriorityWeight := (2.0 / float64(job.Ticket.TicketPriority)) * 2.0

		score := (ageInDays * ageWeight) + deadlineWeight + jobPriorityWeight + ticketPriorityWeight
		scoredJobs[i] = jobWithScore{ID: job.ID, Score: score}
	}

	sort.Slice(scoredJobs, func(i, j int) bool {
		return scoredJobs[i].Score > scoredJobs[j].Score
	})

	for newPriority, scoredJob := range scoredJobs {
		err := j.jobRepo.ForceUpdatePriority(ctx, tx, scoredJob.ID, newPriority+1)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (j *JobReorderJob) getActiveJobsByDepartment(ctx context.Context, tx *sql.Tx, departmentID int) ([]model.Job, error) {
	query := `
        SELECT 
            j.id, j.job_priority,
            t.id as ticket_id, t.created_at, t.ticket_priority, t.deadline
        FROM job j
        JOIN ticket t ON j.ticket_id = t.id
        WHERE t.department_target_id = $1 -- [FIX] Menggunakan kolom dari tabel ticket
        AND EXISTS (
            SELECT 1 FROM track_status_ticket tst
            JOIN status_ticket st ON tst.status_ticket_id = st.id
            WHERE tst.ticket_id = t.id 
            AND tst.finish_date IS NULL
            AND st.name IN ('Menunggu Job', 'Dikerjakan')
        )`

	rows, err := tx.QueryContext(ctx, query, departmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []model.Job
	for rows.Next() {
		var job model.Job
		var ticket model.Ticket
		if err := rows.Scan(
			&job.ID, &job.JobPriority,
			&ticket.ID, &ticket.CreatedAt, &ticket.TicketPriority, &ticket.Deadline,
		); err != nil {
			return nil, err
		}
		job.TicketID = ticket.ID
		job.Ticket = ticket
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (j *JobReorderJob) getActiveTargetDepartments(ctx context.Context) ([]int, error) {
	rows, err := j.db.QueryContext(ctx, "SELECT id FROM department WHERE is_active = true AND receive_job = true")
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
