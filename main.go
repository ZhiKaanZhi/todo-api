package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type Server struct {
	store TodoStorer
}

func (s *Server) handleGetTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := s.store.GetAll(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to fetch todos",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func (s *Server) handleCreateTodo(w http.ResponseWriter, r *http.Request) {
	var newTodo Todo
	err := json.NewDecoder(r.Body).Decode(&newTodo)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid JSON",
		})
		return
	}

	err = s.store.Create(r.Context(), &newTodo)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to create todo",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTodo)
}

func (s *Server) handleDeleteTodo(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid ID type",
		})
		return
	}

	err = s.store.Delete(r.Context(), id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Todo not found",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Todo deleted",
	})
}

func startCleanupWorker(ctx context.Context, pool *pgxpool.Pool, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			result, err := pool.Exec(ctx,
				"DELETE FROM todos WHERE expires_at < NOW()")
			if err != nil {
				log.Printf("Cleanup error: %v", err)
				continue
			}
			if result.RowsAffected() > 0 {
				log.Printf("Cleaned up %d expired todos", result.RowsAffected())
			}
		case <-ctx.Done():
			log.Println("Cleanup worker stopping")
			return
		}
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	dsn := fmt.Sprintf("postgres://%s:%s@localhost:5432/%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PW"),
		os.Getenv("POSTGRES_DB"),
	)

	// 1. Create a context that cancels on Ctrl+C or container stop
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// 2. Create the database connection pool
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	// 3. Build the store and server
	store := &TodoStore{pool: pool}
	srv := &Server{store: store}

	// 4. Start the cleanup worker (runs in background)
	go startCleanupWorker(ctx, pool, 1*time.Minute)

	// 5. Configure HTTP server
	http.HandleFunc("GET /todos", srv.handleGetTodos)
	http.HandleFunc("POST /todos", srv.handleCreateTodo)
	http.HandleFunc("DELETE /todos/{id}", srv.handleDeleteTodo)

	httpServer := &http.Server{
		Addr: ":8080",
	}

	// 6. Start HTTP server in a goroutine
	go func() {
		fmt.Println("TODO API starting on :8080")
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 7. Wait for shutdown signal
	<-ctx.Done()
	log.Println("Shutdown signal received")

	// 8. Drain in-flight requests (5 second timeout)
	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
