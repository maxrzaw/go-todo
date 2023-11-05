package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/maxrzaw/go-todo/models"
	"gorm.io/gorm"
)

type HealthCheckResponse struct {
	Alive    bool   `json:"alive" xml:"alive"`
	Database string `json:"database" xml:"database"`
}

func Healthz(c echo.Context) error {
	status := http.StatusOK
	r := &HealthCheckResponse{
		Alive:    true,
		Database: "healthy",
	}
	var result models.HealthCheck

	models.DB.Where("UUID = ?", models.Hc_uuid).First(&result)

	if result.UUID != models.Hc_uuid {
		r.Database = "unhealthy"
		status = http.StatusServiceUnavailable
	}
	return c.JSON(status, r)
}

func CreateItem(c echo.Context) error {
	t := new(models.TodoItem)
	if err := c.Bind(t); err != nil {
		data := map[string]interface{}{
			"message": err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, data)
	}
	c.Logger().Info("Adding new TodoItem")

	todo := &models.TodoItem{Description: t.Description, Completed: false}
	if err := models.DB.Create(&todo).Error; err != nil {
		data := map[string]interface{}{
			"message": err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, data)
	}

	response := map[string]interface{}{
		"created": todo,
	}

	return c.JSON(http.StatusOK, response)
}

func MarkTodoComplete(c echo.Context) error {
	return MarkItem(c, true)
}

func MarkTodoActive(c echo.Context) error {
	return MarkItem(c, false)
}

func MarkItem(c echo.Context, completed bool) error {
	// Get URL parameter from echo
	id := c.Param("id")
	t := new(models.TodoItem)

	if err := c.Bind(t); err != nil {
		data := map[string]interface{}{
			"message": err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, data)
	}

	todo := new(models.TodoItem)
	if err := models.DB.First(&todo, id).Error; err != nil {
		data := map[string]interface{}{
			"message": err.Error(),
		}

		if err.Error() == "record not found" {
			return c.JSON(http.StatusNotFound, data)
		}

		return c.JSON(http.StatusInternalServerError, data)
	}

	todo.Completed = completed

	if err := models.DB.Save(&todo).Error; err != nil {
		data := map[string]interface{}{
			"message": err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, data)
	}

	return c.JSON(http.StatusOK, todo)
}

func UpdateTodoDescription(c echo.Context) error {
	// Get URL parameter from echo
	id := c.Param("id")
	t := new(models.TodoItem)

	if err := c.Bind(t); err != nil {
		data := map[string]interface{}{
			"message": err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, data)
	}

	existing_todo := new(models.TodoItem)
	if err := models.DB.First(&existing_todo, id).Error; err != nil {
		data := map[string]interface{}{
			"message": err.Error(),
		}

		if err.Error() == "record not found" {
			return c.JSON(http.StatusNotFound, data)
		}

		return c.JSON(http.StatusInternalServerError, data)
	}

	existing_todo.Description = t.Description

	if err := models.DB.Save(&existing_todo).Error; err != nil {
		data := map[string]interface{}{
			"message": err.Error(),
		}

		if err.Error() == "record not found" {
			return c.JSON(http.StatusNotFound, data)
		}

		return c.JSON(http.StatusInternalServerError, data)
	}

	response := map[string]interface{}{
		"data": existing_todo,
	}

	return c.JSON(http.StatusOK, response)
}

func GetTodo(c echo.Context) error {
	// Get URL parameter from echo
	id := c.Param("id")
	c.Logger().Info(id)

	var todos []*models.TodoItem
	if res := models.DB.Debug().Find(&todos, id); res.Error != nil {
		data := map[string]interface{}{
			"message": res.Error.Error(),
		}

		if res.Error.Error() == "record not found" {
			return c.JSON(http.StatusNotFound, data)
		}

		return c.JSON(http.StatusInternalServerError, data)
	}

	response := map[string]interface{}{
		"data": todos[0],
	}

	return c.JSON(http.StatusOK, response)
}

func DeleteTodo(c echo.Context) error {
	// Get URL parameter from echo
	id := c.Param("id")
	t := new(models.TodoItem)

	if err := c.Bind(t); err != nil {
		data := map[string]interface{}{
			"message": err.Error(),
		}

		if err.Error() == "record not found" {
			return c.JSON(http.StatusNotFound, data)
		}

		return c.JSON(http.StatusInternalServerError, data)
	}

	existing_todo := new(models.TodoItem)
	if err := models.DB.First(&existing_todo, id).Error; err != nil {
		data := map[string]interface{}{
			"message": err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, data)
	}

	if err := models.DB.Delete(&existing_todo).Error; err != nil {
		data := map[string]interface{}{
			"message": err.Error(),
		}

		if err.Error() == "record not found" {
			return c.JSON(http.StatusNotFound, data)
		}

		return c.JSON(http.StatusInternalServerError, data)
	}

	response := map[string]interface{}{
		"deleted": "true",
	}

	return c.JSON(http.StatusOK, response)
}

func GetTodos(c echo.Context) error {
	// Get URL parameter from echo
	param := c.QueryParam("completed")

	var todos []*models.TodoItem
	var res *gorm.DB
	if param != "" {
		completed, err := strconv.ParseBool(param)
		if err != nil {
			panic(err)
		}
		res = models.DB.Find(&todos, "completed = ?", completed)
	} else {
		res = models.DB.Find(&todos)
	}
	if res.Error != nil {
		data := map[string]interface{}{
			"message": res.Error.Error(),
		}
		return c.JSON(http.StatusInternalServerError, data)
	}

	return c.JSON(http.StatusOK, todos)
}
