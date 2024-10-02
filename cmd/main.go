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
	dst     = "/tmp"
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
		api.POST("/ping", PongHandler)
		api.POST("/upload", reportHandler)
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

func PongHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, gin.H{
		"message": "pong pong",
	})
}

func reportHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	// single file
	file, _ := c.FormFile("file")
	if file == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "invalid request body",
		})
		return
	}
	log.Println(file.Filename)

	destination := dst + "/" + file.Filename

	// Upload the file to specific dst.
	if err := c.SaveUploadedFile(file, destination); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"etag":   requestid.Get(c),
		"time":   time.Now(),
		"size":   file.Size,
		"file":   fmt.Sprintf("'%s' uploaded!", destination),
	})

}
