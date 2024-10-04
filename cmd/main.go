package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
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
	// Logging to a file.
	logFile := dst + "/api.log"
	f, _ := os.Create(logFile)
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	router := gin.Default()
	router.Use(requestid.New())

	// API routes and handlers
	api := router.Group("/api")
	{
		api.GET("/healthz", healthz)
		api.POST("/ping", PongHandler)
		api.POST("/upload", reportHandler)
		api.POST("/graphql", graphqlHandler)
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

type Hyperlink struct {
	URL  string `json:"url"`
	Verb string `json:"verb"`
}

type UploadDocumentResponse struct {
	Name         string    `json:"name"`
	DocumentID   string    `json:"documentId"`
	ErrorMessage string    `json:"errorMsg"`
	DownloadLink Hyperlink `json:"downloadHyperlink"`
}

func graphqlHandler(c *gin.Context) {
	etag := requestid.Get(c)
	// Get the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}
	err = writeToFile(path.Join(dst, fmt.Sprintf("body-%v.txt", etag)), body)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write headers to file"})
		return
	}

	// Get the request headers
	headers := c.Request.Header

	// Write the headers and body to a file
	err = writeToFile(path.Join(dst, fmt.Sprintf("header-%v.txt", etag)), headers)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write headers to file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"etag":   etag,
		"graphql-response": UploadDocumentResponse{
			Name:       etag,
			DocumentID: etag,
			DownloadLink: Hyperlink{
				URL:  fmt.Sprintf("%s/graphql", dst),
				Verb: "post",
			},
		},
	})
}

func writeToFile(filename string, data interface{}) error {
	var bytes []byte

	switch data.(type) {
	case []byte:
		bytes = data.([]byte)
		log.Printf("%s: %v", filename, string(bytes))
	case http.Header:
		var headerString string
		for key, values := range data.(http.Header) {
			headerString += fmt.Sprintf("%s: %s\n", key, values[0])
		}
		bytes = []byte(headerString)
	default:
		return fmt.Errorf("Unsupported data type: %T", data)
	}

	return os.WriteFile(filename, bytes, 0644)
}
