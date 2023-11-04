package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/maxrzaw/go-todo/models"
	"gorm.io/gorm"
)

func Index(c echo.Context) error {
	var todos []*models.TodoItem
	var res *gorm.DB
	res = models.DB.Find(&todos)
	if res.Error != nil {
		data := map[string]interface{}{
			"message": res.Error.Error(),
		}
		return c.JSON(http.StatusInternalServerError, data)
	}

	return c.Render(http.StatusOK, "index.html", todos)
}

func Todo(c echo.Context) error {
	description := c.FormValue("description")
	c.Logger().Info("Adding new TodoItem with htmx")

	todo := &models.TodoItem{Description: description, Completed: false}
	if err := models.DB.Create(&todo).Error; err != nil {
		data := map[string]interface{}{
			"message": err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, data)
	}

	return c.Render(http.StatusOK, "todo.html", todo)
}
