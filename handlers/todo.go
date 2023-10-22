package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/maxrzaw/go-todo/models"
	"github.com/sirupsen/logrus"
)

type CreateTodoItemRequest struct {
	Description string
}

func GetItemById(Id int) bool {
	todo := &models.TodoItem{}
	result := models.DB.First(&todo, Id)
	if result.Error != nil {
		logrus.Warn("TodoItem not found in database")
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

	logrus.WithFields(logrus.Fields{"Description": body.Description}).Info("Adding new TodoItem.")

	todo := &models.TodoItem{Description: body.Description, Completed: false}
	models.DB.Create(&todo)

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
		logrus.WithFields(logrus.Fields{"Id": id, "Completed": completed}).Info("Updating Completed Status.")

		todo := &models.TodoItem{}
		models.DB.First(&todo, id)
		todo.Completed = true
		models.DB.Save(&todo)

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
		logrus.WithFields(logrus.Fields{"Id": id, "Description": body.Description}).Info("Updating TodoItem")

		todo := &models.TodoItem{}
		models.DB.First(&todo, id)
		todo.Description = body.Description
		models.DB.Save(&todo)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(todo)
	}
}

func GetItem(w http.ResponseWriter, r *http.Request) {
	// Get URL parameter from mux
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	// Test if the TodoItem exist in DB
	todo := &models.TodoItem{}
	result := models.DB.First(&todo, id)
	if result.Error != nil {
		w.WriteHeader(404)
	} else {
		logrus.WithFields(logrus.Fields{"Id": id}).Info("Getting TodoItem")
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
		logrus.WithFields(logrus.Fields{"Id": id}).Info("Deleting TodoItem")
		todo := &models.TodoItem{}
		models.DB.First(&todo, id)
		models.DB.Delete(&todo)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"deleted": true}`)
	}
}

func GetTodoItems(w http.ResponseWriter, r *http.Request) {
	var todos []models.TodoItem
	param := r.URL.Query().Get("completed")
	logrus.WithFields(logrus.Fields{"Completed": param}).Info("Getting Todo Items")
	if param != "" {
		completed, err := strconv.ParseBool(param)
		if err != nil {
			panic(err)
		}
		models.DB.Find(&todos, "completed = ?", completed)
	} else {
		models.DB.Find(&todos)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}
