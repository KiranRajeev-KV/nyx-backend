package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	fmt.Println("Starting server on port 8080...")
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Server is running!",
		})
	})
	router.Run()
}
