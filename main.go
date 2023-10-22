package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var hc_uuid uuid.UUID

type CreateTodoItemRequest struct {
	Description string
}

type TodoItem struct {
	Id          int `gorm:"primary_key"`
	Description string
	Completed   bool
	CreatedAt   time.Time `gorm:"autoCreateTime:true"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime:true"`
}

type HealthCheck struct {
	Id        int `gorm:"primary_key"`
	UUID      uuid.UUID
	UpdatedAt time.Time `gorm:"autoUpdateTime:true"`
}

func Healthz(w http.ResponseWriter, r *http.Request) {
	log.Info("healthcheck called")
	w.Header().Set("Content-Type", "application/json")
	db_connection := "healthy"
	var result HealthCheck

	db.Where("UUID = ?", hc_uuid).First(&result)

	if result.UUID != hc_uuid {
		db_connection = "unhealthy"
		w.WriteHeader(400)
	}

	response := fmt.Sprintf(`{"alive": true, "database": "%s"}`, db_connection)
	io.WriteString(w, response)
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetReportCaller(true)
}

func GetItemById(Id int) bool {
	todo := &TodoItem{}
	result := db.First(&todo, Id)
	if result.Error != nil {
		log.Warn("TodoItem not found in database")
		return false
	}
	return true
}

func CreateItem(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var body CreateTodoItemRequest
	err := decoder.Decode(&body)
	if err != nil {
		panic(err)
	}

	log.WithFields(log.Fields{"Description": body.Description}).Info("Adding new TodoItem.")

	todo := &TodoItem{Description: body.Description, Completed: false}
	db.Create(&todo)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func MarkItemAsComplete(w http.ResponseWriter, r *http.Request) {
	MarkItem(w, r, true)
}

func MarkItemAsIncomplete(w http.ResponseWriter, r *http.Request) {
	MarkItem(w, r, false)
}

func MarkItem(w http.ResponseWriter, r *http.Request, completed bool) {
	// Get URL parameter from mux
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	exists := GetItemById(id)
	if exists == false {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		io.WriteString(w, `{"updated": false, "error": "Record Not Found"}`)
	} else {
		log.WithFields(log.Fields{"Id": id, "Completed": completed}).Info("Updating Completed Status.")

		todo := &TodoItem{}
		db.First(&todo, id)
		todo.Completed = true
		db.Save(&todo)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(todo)
	}
}

func UpdateItemDescription(w http.ResponseWriter, r *http.Request) {
	// Get URL parameter from mux
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	decoder := json.NewDecoder(r.Body)
	var body CreateTodoItemRequest
	err := decoder.Decode(&body)
	if err != nil {
		panic(err)
	}

	// Test if the TodoItem exist in DB
	exists := GetItemById(id)
	if exists == false {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"updated": false, "error": "Record Not Found"}`)
	} else {
		log.WithFields(log.Fields{"Id": id, "Description": body.Description}).Info("Updating TodoItem")

		todo := &TodoItem{}
		db.First(&todo, id)
		todo.Description = body.Description
		db.Save(&todo)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(todo)
	}
}

func GetItem(w http.ResponseWriter, r *http.Request) {
	// Get URL parameter from mux
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	// Test if the TodoItem exist in DB
	todo := &TodoItem{}
	result := db.First(&todo, id)
	if result.Error != nil {
		w.WriteHeader(404)
	} else {
		log.WithFields(log.Fields{"Id": id}).Info("Getting TodoItem")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(todo)
	}
}

func DeleteItem(w http.ResponseWriter, r *http.Request) {
	// Get URL parameter from mux
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	// Test if the TodoItem exist in DB
	exists := GetItemById(id)
	if exists == false {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		io.WriteString(w, `{"deleted": false, "error": "Record Not Found"}`)
	} else {
		log.WithFields(log.Fields{"Id": id}).Info("Deleting TodoItem")
		todo := &TodoItem{}
		db.First(&todo, id)
		db.Delete(&todo)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"deleted": true}`)
	}
}

func GetTodoItems(w http.ResponseWriter, r *http.Request) {
	var todos []TodoItem
	param := r.URL.Query().Get("completed")
	log.WithFields(log.Fields{"Completed": param}).Info("Getting Todo Items")
	if param != "" {
		completed, err := strconv.ParseBool(param)
		if err != nil {
			panic(err)
		}
		db.Find(&todos, "completed = ?", completed)
	} else {
		db.Find(&todos)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func InitDB() {
	dsn := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable TimeZone=%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DATABASE"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_TZ"),
	)
	log.Info(dsn)

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Error(err)
		panic("failed to connect database")
	}

	db.Debug().Migrator().AutoMigrate(&HealthCheck{})
	hc_uuid = uuid.New()
	hcm := HealthCheck{UUID: hc_uuid}
	db.Create(&hcm)
}

func main() {
	InitDB()
	// Migrate the schema
	if db.Debug().Migrator().HasTable(&TodoItem{}) {
		// I don't think this is needed
		// db.Debug().Migrator().DropTable(&TodoItemModel{})
	}
	db.Debug().Migrator().AutoMigrate(&TodoItem{})

	log.Info("Starting API Server")
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
	log.Fatal(http.ListenAndServe(":80", handler))
}
