package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Mock store — implements TodoStorer without a database.
type MockStore struct {
	todos  map[int]Todo
	nextID int
}

func NewMockStore(todos map[int]Todo) *MockStore {
	return &MockStore{
		todos:  todos,
		nextID: len(todos) + 1,
	}
}

func (m *MockStore) GetAll(ctx context.Context) ([]Todo, error) {
	result := make([]Todo, 0, len(m.todos))
	for _, t := range m.todos {
		result = append(result, t)
	}
	return result, nil
}

func (m *MockStore) Create(ctx context.Context, t *Todo) error {
	t.ID = m.nextID
	t.CreatedAt = time.Now()
	t.ExpiresAt = time.Now().Add(72 * time.Hour)
	m.todos[m.nextID] = *t
	m.nextID++
	return nil
}

func (m *MockStore) Delete(ctx context.Context, id int) error {
	if _, exists := m.todos[id]; !exists {
		return fmt.Errorf("todo not found")
	}
	delete(m.todos, id)
	return nil
}

// Tests

func TestHandleGetTodos(t *testing.T) {
	tests := []struct {
		name           string
		setupTodos     map[int]Todo
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "returns all todos",
			setupTodos: map[int]Todo{
				1: {ID: 1, Title: "First"},
				2: {ID: 2, Title: "Second"},
			},
			expectedStatus: 200,
			expectedCount:  2,
		},
		{
			name:           "returns empty when no todos",
			setupTodos:     map[int]Todo{},
			expectedStatus: 200,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{store: NewMockStore(tt.setupTodos)}

			req := httptest.NewRequest("GET", "/todos", nil)
			rec := httptest.NewRecorder()

			srv.handleGetTodos(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("status: got %d, want %d", rec.Code, tt.expectedStatus)
			}

			var got []Todo
			json.NewDecoder(rec.Body).Decode(&got)
			if len(got) != tt.expectedCount {
				t.Errorf("count: got %d, want %d", len(got), tt.expectedCount)
			}
		})
	}
}

func TestHandleCreateTodo(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name:           "create new todo",
			body:           `{"title":"Test","description":"A test todo"}`,
			expectedStatus: 201,
		},
		{
			name:           "garbage body",
			body:           `this is not json`,
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{store: NewMockStore(map[int]Todo{})}

			body := strings.NewReader(tt.body)
			req := httptest.NewRequest("POST", "/todos", body)
			rec := httptest.NewRecorder()

			srv.handleCreateTodo(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("status: got %d, want %d", rec.Code, tt.expectedStatus)
			}

			if tt.expectedStatus == 201 {
				var got Todo
				json.NewDecoder(rec.Body).Decode(&got)
				if got.ID == 0 {
					t.Error("expected todo to have an ID assigned")
				}
				if got.Title != "Test" {
					t.Errorf("title: got %q, want %q", got.Title, "Test")
				}
			}
		})
	}
}

func TestHandleDeleteTodo(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		setupTodos     map[int]Todo
		expectedStatus int
	}{
		{
			name: "delete existing todo",
			id:   "1",
			setupTodos: map[int]Todo{
				1: {ID: 1, Title: "First"},
				2: {ID: 2, Title: "Second"},
			},
			expectedStatus: 200,
		},
		{
			name:           "garbage id",
			id:             "garbage",
			setupTodos:     map[int]Todo{},
			expectedStatus: 400,
		},
		{
			name:           "unknown id",
			id:             "3",
			setupTodos:     map[int]Todo{},
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{store: NewMockStore(tt.setupTodos)}

			req := httptest.NewRequest("DELETE", "/todos/"+tt.id, nil)
			req.SetPathValue("id", tt.id)
			rec := httptest.NewRecorder()

			srv.handleDeleteTodo(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("status: got %d, want %d", rec.Code, tt.expectedStatus)
			}
		})
	}
}
