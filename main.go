package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

const (
	appPort = "8080"
)

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = appPort
	}

	svc := setupRouter()

	if err := svc.Run(fmt.Sprintf(":%v", port)); err != nil {
		log.Fatal(err)
	}
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.Use(requestid.New())

	// API routes and handlers
	api := router.Group("/api")
	{
		api.GET("/healthz", healthz)
		//api.POST("/clusters/:cluserId", UpdateClusterHandler)
	}

	return router
}

func healthz(c *gin.Context) {

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, gin.H{
		"status": "ok:",
		"etag":   requestid.Get(c),
		"time":   time.Now(),
	})
}
