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
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var hc_uuid uuid.UUID

type CreateTodoItemRequest struct {
	Description string
}

type TodoItemModel struct {
	Id          int `gorm:"primary_key"`
	Description string
	Completed   bool
	CreatedAt   time.Time `gorm:"autoCreateTime:true"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime:true"`
}

type HealthCheckModel struct {
	Id        int `gorm:"primary_key"`
	UUID      uuid.UUID
	UpdatedAt time.Time `gorm:"autoUpdateTime:true"`
}

func Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	db_connection := "healthy"
	var result HealthCheckModel

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
	todo := &TodoItemModel{}
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

	todo := &TodoItemModel{Description: body.Description, Completed: false}
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

		todo := &TodoItemModel{}
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

		todo := &TodoItemModel{}
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
	todo := &TodoItemModel{}
	result := db.First(&todo, id)
	if result.Error != nil {
		w.WriteHeader(404)
	} else {
		log.WithFields(log.Fields{"Id": id}).Info("Deleting TodoItem")
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
		todo := &TodoItemModel{}
		db.First(&todo, id)
		db.Delete(&todo)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"deleted": true}`)
	}
}

func GetTodoItems(w http.ResponseWriter, r *http.Request) {
	var todos []TodoItemModel
	param := r.URL.Query().Get("completed")
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
		"%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_DATABASE"),
	)
	log.Info(dsn)

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Error(err)
		panic("failed to connect database")
	}

	db.Debug().Migrator().AutoMigrate(&HealthCheckModel{})
	hc_uuid = uuid.New()
	hcm := HealthCheckModel{UUID: hc_uuid}
	db.Create(&hcm)
}

func main() {
	InitDB()
	// Migrate the schema
	if db.Debug().Migrator().HasTable(&TodoItemModel{}) {
		// I don't think this is needed
		// db.Debug().Migrator().DropTable(&TodoItemModel{})
	}
	db.Debug().Migrator().AutoMigrate(&TodoItemModel{})

	log.Info("Starting API Server")
	router := mux.NewRouter()

	// Health Check
	router.HandleFunc("/healthz", Healthz).Methods("GET")

	// /todo
	router.HandleFunc("/todo", CreateItem).Methods("POST")

	// /todo/{id}
	router.HandleFunc("/todo/{id}", GetItem).Methods("GET")
	router.HandleFunc("/todo/{id}", DeleteItem).Methods("DELETE")
	router.HandleFunc("/todo/{id}/update-description", UpdateItemDescription).Methods("POST")
	router.HandleFunc("/todo/{id}/mark-complete", MarkItemAsComplete).Methods("POST")
	router.HandleFunc("/todo/{id}/mark-incomplete", MarkItemAsIncomplete).Methods("POST")

	// /todo/list
	router.HandleFunc("/todo/list", GetTodoItems).Methods("GET")

	log.Fatal(http.ListenAndServe(":80", router))
}
