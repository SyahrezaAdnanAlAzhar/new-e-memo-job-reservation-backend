package main

import (
	"database/sql"
	"log"
	"os"

	"e-memo-job-reservation-api/pkg/database"
)

func main() {
	db := database.Connect()
	defer db.Close()

	log.Println("Starting migration: Create websocket_tickets table...")

	migration := `
-- Create websocket_tickets table if it doesn't exist
-- This table stores temporary tickets for WebSocket connections
-- Supports both authenticated users (with user_id) and public/anonymous connections (user_id = NULL)

CREATE TABLE IF NOT EXISTS public.websocket_tickets (
    ticket TEXT PRIMARY KEY,
    user_id BIGINT,  -- Nullable to support public/anonymous WebSocket connections
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Add foreign key constraint if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'websocket_tickets_user_id_fkey'
    ) THEN
        ALTER TABLE public.websocket_tickets
        ADD CONSTRAINT websocket_tickets_user_id_fkey 
        FOREIGN KEY (user_id) REFERENCES public.app_user(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Add index on expires_at for cleanup operations
CREATE INDEX IF NOT EXISTS idx_websocket_tickets_expires_at 
ON public.websocket_tickets(expires_at);

-- Add index on user_id for user lookup
CREATE INDEX IF NOT EXISTS idx_websocket_tickets_user_id 
ON public.websocket_tickets(user_id);

-- Add comment to explain the nullable user_id
COMMENT ON COLUMN public.websocket_tickets.user_id IS 'User ID for authenticated users, NULL for public/anonymous connections';
COMMENT ON TABLE public.websocket_tickets IS 'Temporary tickets for WebSocket connections. Supports both authenticated and anonymous users.';
`

	// Execute migration
	_, err := db.Exec(migration)
	if err != nil {
		log.Fatalf("Failed to run migration: %v", err)
	}

	log.Println("Migration completed successfully!")

	// Verify table exists
	var tableName string
	err = db.QueryRow("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'websocket_tickets'").Scan(&tableName)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Warning: Table websocket_tickets was not found after migration")
		} else {
			log.Printf("Error verifying table: %v", err)
		}
	} else {
		log.Printf("Table '%s' exists and is ready to use", tableName)
	}

	os.Exit(0)
}
