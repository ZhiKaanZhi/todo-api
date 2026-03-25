package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func handleGetTodos(w http.ResponseWriter, r *http.Request) {
	// w is what you write the response TO (like Response in C#)
	// r is the incoming request FROM the client (like Request in C#)
	// customString := fmt.Sprintf("ID = %d\nTitle = %s\nDesciption = %s\nDone= %b\nCreatedAt=%t\nUpdatedAt = %t", todo.ID, todo.Title, todo.Description, todo.Done, todo.CreatedAt, todo.UpdatedAt)
	json.NewEncoder(w).Encode(todos)
}

func handleCreateTodo(w http.ResponseWriter, r *http.Request) {
	var newTodo Todo
	err := json.NewDecoder(r.Body).Decode(&newTodo)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest) // 400
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid JSON",
		})
		return
	}

	// Create new todo
	newTodo.ID = nextID
	newTodo.CreatedAt = time.Now()
	newTodo.Done = false
	todos[nextID] = newTodo
	nextID++

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 instead of 200
	json.NewEncoder(w).Encode(newTodo)
}

func handleDeleteTodo(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	id, err := strconv.Atoi(idString)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest) // 400
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid ID type",
		})
		return
	}

	_, exists := todos[id]
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound) // 404
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Todo not found",
		})
		return
	}

	delete(todos, id)

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Todo deleted",
	})

}

func main() {

	todos[1] = Todo{
		ID:          1,
		Title:       "Practice Go",
		Description: "First project",
		Done:        false,
		CreatedAt:   time.Now(),
	}
	todos[2] = Todo{
		ID:          2,
		Title:       "Practice Advanced Go",
		Description: "Expand the first project",
		Done:        false,
		CreatedAt:   time.Now(),
	}

	nextID = 3

	http.HandleFunc("GET /todos", handleGetTodos)
	http.HandleFunc("POST /todos", handleCreateTodo)
	http.HandleFunc("DELETE /todos/{id}", handleDeleteTodo)

	fmt.Println("TODO API starting on :8080")
	http.ListenAndServe(":8080", nil)
}
