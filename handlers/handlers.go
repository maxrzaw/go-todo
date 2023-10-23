package handlers

import (
	"github.com/labstack/echo/v4"
)

func AddHandlers(e *echo.Echo) {
	e.Logger.Warn("Adding handlers")

	api := e.Group("/api")
	api.GET("/healthz", Healthz)

	todos := api.Group("/todos")
	todos.GET("/list", GetTodos)

	todo := todos.Group("/todo")
	todo.POST("", CreateItem)
	todo.GET("/:id", GetTodo)
	todo.DELETE("/:id", DeleteTodo)

	todo.POST("/:id/update-description", UpdateTodoDescription)
	todo.POST("/:id/mark-complete", MarkTodoComplete)
	todo.POST("/:id/mark-incomplete", MarkTodoIncomplete)
}
