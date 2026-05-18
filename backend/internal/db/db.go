package db

import (
	"fmt"
	"log"

	"github.com/example/worklog-system/internal/config"
	"github.com/example/worklog-system/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect db failed: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{}, &models.Project{}, &models.Timesheet{}, &models.BackfillLog{},
		&models.WorkPlan{}, &models.WorkScheduleTask{}, &models.WorkScheduleDay{},
	); err != nil {
		log.Fatalf("automigrate failed: %v", err)
	}
	return db
}
