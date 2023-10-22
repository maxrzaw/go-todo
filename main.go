package main

import (
	"net/http"

	"github.com/maxrzaw/go-todo/handlers"
	"github.com/maxrzaw/go-todo/models"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetReportCaller(true)
}

func main() {
	models.InitDb()

	logrus.Info("Starting API Server")
	handler := handlers.Init()
	logrus.Fatal(http.ListenAndServe(":80", handler))
}
