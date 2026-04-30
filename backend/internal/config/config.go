package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

type App struct {
	Config *oauth2.Config
}

type DBConfig struct {
	DBDriver string
	DBSource string
}

func LoadDBConfig() DBConfig {
	err := godotenv.Load() // Load environment variables from .env file
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	return DBConfig{
		DBDriver: "postgres",
		DBSource: os.Getenv("DBSource"),
	}
}
