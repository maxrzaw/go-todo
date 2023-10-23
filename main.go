package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/maxrzaw/go-todo/handlers"
	"github.com/maxrzaw/go-todo/models"
)

func main() {
	models.InitDb()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	e.Use(middleware.CORS())
	handlers.AddHandlers(e)
	e.Logger.Fatal(e.Start(":80"))
}
