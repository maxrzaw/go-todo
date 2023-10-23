package handlers

import (
	"io"
	"net/http"
	"text/template"

	"github.com/labstack/echo/v4"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func AddHandlers(e *echo.Echo) {
	e.Logger.Warn("Adding handlers")

	t := &Template{
		templates: template.Must(template.ParseGlob("handlers/templates/*.html")),
	}

	e.Renderer = t

	e.GET("index.html", Index)
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

func Index(c echo.Context) error {
	return c.Render(http.StatusOK, "index", map[string]interface{}{})
}
