package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

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
			// 1. Set up: put test data in the global todos map
			todos = tt.setupTodos

			// 2. Create a fake request with httptest
			req := httptest.NewRequest("GET", "/todos", nil)
			rec := httptest.NewRecorder()
			// 3. Call the handler
			handleGetTodos(rec, req)
			// 4. Check the status code and response body
			if rec.Code != tt.expectedStatus {
				t.Errorf("status: got %d, want %d", rec.Code, tt.expectedStatus)
			}
		})
	}
}

func TestHandleCreateTodos(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setupTodos     map[int]Todo
		expectedStatus int
	}{
		{
			name:           "create new todo",
			body:           `{"title":"Test","description":"A test todo"}`,
			setupTodos:     map[int]Todo{},
			expectedStatus: 201,
		},
		{
			name:           "garbage body",
			body:           `this is not json`,
			setupTodos:     map[int]Todo{},
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Set up: put test data in the global todos map
			todos = tt.setupTodos
			nextID = 1

			// 2. Create a fake request with httptest
			body := strings.NewReader(tt.body)
			req := httptest.NewRequest("POST", "/todos", body)
			rec := httptest.NewRecorder()
			// 3. Call the handler
			handleCreateTodo(rec, req)
			// 4. Check the status code and response body
			if rec.Code != tt.expectedStatus {
				t.Errorf("status: got %d, want %d", rec.Code, tt.expectedStatus)
			}
		})
	}
}

func TestHandleDeleteTodos(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		setupTodos     map[int]Todo
		expectedStatus int
	}{
		{
			name: "delete todo",
			id:   `1`,
			setupTodos: map[int]Todo{
				1: {ID: 1, Title: "First"},
				2: {ID: 2, Title: "Second"},
			},
			expectedStatus: 200,
		},
		{
			name:           "garbage id",
			id:             `garbage`,
			setupTodos:     map[int]Todo{},
			expectedStatus: 400,
		},
		{
			name:           "unknown id",
			id:             `3`,
			setupTodos:     map[int]Todo{},
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Set up: put test data in the global todos map
			todos = tt.setupTodos

			// 2. Create a fake request with httptest
			req := httptest.NewRequest("DELETE", "/todos/"+tt.id, nil)
			req.SetPathValue("id", tt.id)
			rec := httptest.NewRecorder()
			// 3. Call the handler
			handleDeleteTodo(rec, req)
			// 4. Check the status code and response body
			if rec.Code != tt.expectedStatus {
				t.Errorf("status: got %d, want %d", rec.Code, tt.expectedStatus)
			}
		})
	}
}
