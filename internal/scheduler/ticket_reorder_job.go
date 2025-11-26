package scheduler

import (
	"context"
	"database/sql"
	"log"
	"math"
	"sort"
	"time"

	"e-memo-job-reservation-api/internal/model"
	"e-memo-job-reservation-api/internal/repository"
	"e-memo-job-reservation-api/internal/websocket"

	"github.com/gin-gonic/gin"
)

type ticketWithScore struct {
	ID    int
	Score float64
}

type TicketReorderJob struct {
	ticketRepo *repository.TicketRepository
	db         *sql.DB
	hub        *websocket.Hub
}

func NewTicketReorderJob(db *sql.DB, ticketRepo *repository.TicketRepository, hub *websocket.Hub) *TicketReorderJob {
	return &TicketReorderJob{db: db, ticketRepo: ticketRepo, hub: hub}
}

// RUN
func (j *TicketReorderJob) Run() {
	log.Println("Starting ticket priority recalculation job...")

	ctx := context.Background()

	departmentIDs, err := j.getActiveTargetDepartments(ctx)
	if err != nil {
		log.Printf("ERROR: Could not get target departments: %v", err)
		return
	}

	for _, deptID := range departmentIDs {
		log.Printf("Processing department ID: %d", deptID)
		err := j.reorderTicketsForDepartment(ctx, deptID)
		if err != nil {
			log.Printf("ERROR: Failed to reorder tickets for department %d: %v", deptID, err)
			continue
		}
	}

	payload := gin.H{"message": "Ticket priorities have been recalculated by the system."}
	message, err := websocket.NewMessage("TICKET_PRIORITY_RECALCULATED", payload)
	if err != nil {
		log.Printf("CRITICAL: Failed to create websocket message for ticket cron job: %v", err)
	} else {
		j.hub.BroadcastMessage(message)
	}

	log.Println("Ticket priority recalculation job finished.")
}

func (j *TicketReorderJob) reorderTicketsForDepartment(ctx context.Context, departmentID int) error {
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tickets, err := j.getActiveTicketsByDepartment(ctx, tx, departmentID)
	if err != nil {
		return err
	}
	if len(tickets) == 0 {
		log.Printf("No active tickets for department ID: %d. Skipping.", departmentID)
		return nil
	}

	// SCORING FOR EACH TICKET
	scoredTickets := make([]ticketWithScore, len(tickets))
	for i, ticket := range tickets {
		ageInDays := time.Since(ticket.CreatedAt).Hours() / 24

		// WEIGHT LOGIC
		ageWeight := calculateAgeWeight(ageInDays)
		priorityWeight := 2.0 / float64(ticket.TicketPriority)

		// DEADLINE WEIGHT
		deadlineWeight := calculateDeadlineWeight(ticket.Deadline)

		score := (ageInDays * ageWeight * 1.0) + (priorityWeight * 1.5) + (deadlineWeight * 2.0)

		scoredTickets[i] = ticketWithScore{ID: ticket.ID, Score: score}
	}

	// SORT TICKET BASED ON SCORE (DESCENDING)
	sort.Slice(scoredTickets, func(i, j int) bool {
		return scoredTickets[i].Score > scoredTickets[j].Score
	})

	// UPDATE PRIORITY
	for newPriority, scoredTicket := range scoredTickets {
		err := j.ticketRepo.ForceUpdatePriority(ctx, tx, scoredTicket.ID, newPriority+1)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func calculateAgeWeight(days float64) float64 {
	if days <= 7 {
		return 1.0
	}
	if days <= 14 {
		return 1.5
	}
	// EXAMPLE
	// (math.Sqrt(14) * 0.5) ~= 1.87
	// (math.Sqrt(21) * 0.5) ~= 2.29
	// (math.Sqrt(28) * 0.5) ~= 2.64
	return math.Sqrt(days) * 0.5
}

func calculateDeadlineWeight(deadline sql.NullTime) float64 {
	if !deadline.Valid {
		return 10.0
	}

	daysRemaining := time.Until(deadline.Time).Hours() / 24

	const steepnessFactor = 3.0
	const baseScore = 100.0
	const minScore = 5.0

	if daysRemaining >= 0 {
		weight := (baseScore-minScore)*math.Exp(-daysRemaining/steepnessFactor) + minScore
		return weight
	} else {
		const penaltyPerDay = 15.0
		return baseScore + (-daysRemaining * penaltyPerDay)
	}
}

// HELPER

// GET ACTIVE TARGET DEPARMENT
func (j *TicketReorderJob) getActiveTargetDepartments(ctx context.Context) ([]int, error) {
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

// GET ALL TICKET ACTIVE FROM SELECTED TARGET DEPARTMENT
func (j *TicketReorderJob) getActiveTicketsByDepartment(ctx context.Context, tx *sql.Tx, departmentID int) ([]model.Ticket, error) {
	query := `
        SELECT t.id, t.created_at, t.ticket_priority, t.deadline
        FROM ticket t
        WHERE t.department_target_id = $1
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

	var tickets []model.Ticket
	for rows.Next() {
		var t model.Ticket
		if err := rows.Scan(&t.ID, &t.CreatedAt, &t.TicketPriority, &t.Deadline); err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}
	return tickets, nil
}
