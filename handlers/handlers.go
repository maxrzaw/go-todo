package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func Init() http.Handler {
	router := mux.NewRouter()

	// Health Check
	router.HandleFunc("/api/healthz", Healthz).Methods("GET")

	// /todo
	router.HandleFunc("/api/todos/todo", CreateItem).Methods("POST")

	// /todos
	router.HandleFunc("/api/todos/list", GetTodoItems).Methods("GET")

	// /todo/{id}
	router.HandleFunc("/api/todos/todo/{id}", GetItem).Methods("GET")
	router.HandleFunc("/api/todos/todo/{id}", DeleteItem).Methods("DELETE")
	router.HandleFunc("/api/todos/todo/{id}/update-description", UpdateItemDescription).Methods("POST")
	router.HandleFunc("/api/todos/todo/{id}/mark-complete", MarkItemAsComplete).Methods("POST")
	router.HandleFunc("/api/todos/todo/{id}/mark-incomplete", MarkItemAsIncomplete).Methods("POST")

	handler := cors.New(
		cors.Options{AllowedMethods: []string{"GET", "POST", "DELETE"}},
	).Handler(router)
	return handler
}
