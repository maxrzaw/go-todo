package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/maxrzaw/go-todo/handlers"
	"github.com/maxrzaw/go-todo/models"
	"github.com/maxrzaw/go-todo/template"
)

func main() {
	env := os.Getenv("TODO_ENV")
	if env == "" {
		env = "local"
	}
	err := godotenv.Load(".env." + env)
	if err != nil {
		log.Fatal("Error loading .env.env file")
	}
	godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	models.InitDb()

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	e.Use(middleware.CORS())

	// Tailwind files
	e.Static("/dist", "dist")

	template.NewTemplateRenderer(e,
		template.TemplateRecipe{
			Name:  "todo.html",
			Base:  "todo.html",
			Paths: []string{"public/todo.html"},
		},
		template.TemplateRecipe{
			Name:  "index.html",
			Base:  "base.html",
			Paths: []string{"public/index.html", "public/todo.html", "public/base.html"},
		},
	)

	handlers.AddHandlers(e)

	e.Logger.Fatal(e.Start(":8080"))
}
