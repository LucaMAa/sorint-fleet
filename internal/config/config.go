package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"sorint-fleet/internal/model"
)

var DB *gorm.DB

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func LoadConfig() *DatabaseConfig {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  File .env non trovato, utilizzo variabili d'ambiente")
	}

	config := &DatabaseConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}

	if config.Host == "" {
		log.Fatal("DB_HOST missing")
	}
	if config.Port == "" {
		log.Fatal("DB_PORT missing")
	}
	if config.User == "" {
		log.Fatal("DB_USER missing")
	}
	if config.Password == "" {
		log.Fatal("DB_PASSWORD missing")
	}
	if config.Name == "" {
		log.Fatal("DB_NAME missing")
	}
	if config.SSLMode == "" {
		log.Fatal("DB_SSLMODE missing")
	}

	return config
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
}

func InitDB(config *DatabaseConfig) {
	var err error
	DB, err = gorm.Open(postgres.Open(config.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("error during connection to database: %v", err)
	}

	if err := DB.AutoMigrate(
		&model.User{},
		&model.Vehicle{},
		&model.Brand{},
		&model.Model{},
		&model.VehicleAssignment{},
	); err != nil {
		log.Fatalf("error during automigrate: %v", err)
	}
}
