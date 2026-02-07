package main

import (
	"context"
	"fmt"
	"juvens-library/internal/routes"
	"log/slog"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()
	oauthCfg := &oauth2.Config{}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8848"
	}
	addr := fmt.Sprint(":", port)
	router := routes.CreateRouter()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
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
