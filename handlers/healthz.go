package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/maxrzaw/go-todo/models"
	"github.com/sirupsen/logrus"
)

func Healthz(w http.ResponseWriter, r *http.Request) {
	logrus.Info("healthcheck called")
	w.Header().Set("Content-Type", "application/json")
	db_connection := "healthy"
	var result models.HealthCheck

	models.DB.Where("UUID = ?", models.Hc_uuid).First(&result)

	if result.UUID != models.Hc_uuid {
		db_connection = "unhealthy"
		w.WriteHeader(400)
	}

	response := fmt.Sprintf(`{"alive": true, "database": "%s"}`, db_connection)
	io.WriteString(w, response)
}
