package main

import (
	"fmt"
	"juvens-library/internal/database"
	"juvens-library/internal/routes"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8848"
	}
	addr := fmt.Sprint(":", port)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	dbObj, err := database.OpenDB()
	if err != nil {
		logger.Error("Failed to open database", "error", err)
	} else {
		logger.Info("Database connection established")
	}
	defer func() {
		if err := database.CloseDB(dbObj); err != nil {
			logger.Error("Failed to close database", "error", err)
		}
	}()

	router := routes.CreateRouter(dbObj)
	//listenandserve only returns error, thus unless the server crashes or we shut it, this wont be
	//displayed if its after the func
	logger.Info("Server starting", "address", addr)
	server := http.Server{
		Addr:    addr, //host:8848
		Handler: router,
	}
	if err := server.ListenAndServe(); err != nil {
		logger.Error("Server failed", "error", err)
		os.Exit(1)
	}

}
