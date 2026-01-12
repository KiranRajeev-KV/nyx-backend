package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KiranRajeev-KV/nyx-backend/cmd"
	"github.com/gin-gonic/gin"
)

func StartServer() {
	router := gin.Default()

	fmt.Println("Starting server on port 8080...")
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Server is running!",
		})
	})

	env, err := cmd.LoadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	server := &http.Server{
		Addr:    ":" + env.Port,
		Handler: router,
	}

	go func() {
		fmt.Println("[OK]: Start the server on port " + ":" + env.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("could not listen on port "+":"+env.Port, err)
		} // Blocking in nature
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("[OK]: Shutting down server...")

	// 10 seconds timeout for the server to shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Println("[OK]: Server forced to shutdown", err)
	}
}

func main() {
	StartServer()
}
