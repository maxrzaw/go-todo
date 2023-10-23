package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/maxrzaw/go-todo/models"
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
