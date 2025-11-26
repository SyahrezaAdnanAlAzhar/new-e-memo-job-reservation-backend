package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"e-memo-job-reservation-api/internal/repository"
	"e-memo-job-reservation-api/internal/scheduler"
	"e-memo-job-reservation-api/internal/websocket"
	"e-memo-job-reservation-api/pkg/database"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	// INIT CONNECTION
	db := database.Connect()
	defer db.Close()

	authRepo := repository.NewAuthRepository(db)
	ticketRepo := repository.NewTicketRepository(db)
	jobRepo := repository.NewJobRepository(db)

	hub := websocket.NewHub(authRepo)
	go hub.Run()

	// CREATE INSTANCE
	ticketReorderJob := scheduler.NewTicketReorderJob(db, ticketRepo, hub)
	jobReorderJob := scheduler.NewJobReorderJob(db, jobRepo, hub)

	// INIT SCHEDULER
	jakartaLocation, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Fatalf("Could not load location Asia/Jakarta: %v", err)
	}
	c := cron.New(cron.WithLocation(jakartaLocation))

	c.AddJob("*/30 * * * *", ticketReorderJob)
	c.AddJob("1-59/30 * * * *", jobReorderJob)

	c.Start()
	log.Println("Cron job scheduler started.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("Shutting down scheduler...")
	ctx := c.Stop()
	<-ctx.Done()
	log.Println("Scheduler gracefully stopped.")
}
