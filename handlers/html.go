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
	res = models.DB.Find(&todos).Order("updated_at asc")
	if res.Error != nil {
		data := map[string]interface{}{
			"message": res.Error.Error(),
		}
		return c.JSON(http.StatusInternalServerError, data)
	}

	return c.Render(http.StatusOK, "index.html", todos)
}

func RemoveTodo(c echo.Context) error {
	c.Logger().Info("Removing TodoItem with htmx")

	// Get URL parameter from echo
	id := c.Param("id")
	t := new(models.TodoItem)

	if err := c.Bind(t); err != nil {
		data := map[string]interface{}{
			"message": err.Error(),
		}

		if err.Error() == "record not found" {
			return c.JSON(http.StatusNoContent, data)
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
			return c.JSON(http.StatusNoContent, data)
		}

		return c.JSON(http.StatusInternalServerError, data)
	}
	return c.NoContent(http.StatusOK)
}

func CompleteTodo(c echo.Context) error {
	return MarkTodo(c, true)
}

func ActiveTodo(c echo.Context) error {
	return MarkTodo(c, false)
}

func MarkTodo(c echo.Context, completed bool) error {
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

	// if completed {
	// 	c.Response().Header().Set("HX-Retarget", "#completed-todos")
	// } else {
	// 	c.Response().Header().Set("HX-Retarget", "#active-todos")
	// }

	// c.Response().Header().Set("HX-Reswap", "beforeend")
	return c.Render(http.StatusOK, "move-todo.html", todo)
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
