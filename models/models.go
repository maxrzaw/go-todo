package models

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
var Hc_uuid uuid.UUID

type TodoItem struct {
	Id          int `gorm:"primary_key"`
	Description string
	Completed   bool
	CreatedAt   time.Time `gorm:"autoCreateTime:true"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime:true"`
}

type HealthCheck struct {
	Id        int `gorm:"primary_key"`
	UUID      uuid.UUID
	UpdatedAt time.Time `gorm:"autoUpdateTime:true"`
}

func InitDb() {
	dsn := fmt.Sprintf(
		"user=%s password=%s m.DBname=%s host=%s port=%s sslmode=disable TimeZone=%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DATABASE"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_TZ"),
	)
	log.Info(dsn)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Error(err)
		panic("failed to connect database")
	}

	DB.Debug().Migrator().AutoMigrate(&HealthCheck{})
	Hc_uuid = uuid.New()
	hcm := HealthCheck{UUID: Hc_uuid}
	DB.Create(&hcm)
	DB.Debug().Migrator().AutoMigrate(&TodoItem{})
}
