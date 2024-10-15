package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/rashedkvm/bct-service/pkg/graphql"
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
		api.POST("/local/graphql", graphqlLocalHandler)
		api.POST("/data", documentHandler)
		api.POST("/data/document/upload", documentHandler)
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

func documentHandler(c *gin.Context) {
	log.Println(c.Request.URL)
	c.Header("Content-Type", "application/json")
	requestContentType := c.Request.Header.Get("Content-Type")

	multipartFormData := strings.Split(requestContentType, ";")

	if len(multipartFormData) > 0 && multipartFormData[0] == "multipart/form-data" {
		log.Println(multipartFormData)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func graphqlHandler(c *gin.Context) {
	etag := requestid.Get(c)
	writeError := writeHttpRequestToFile(path.Join(dst, fmt.Sprintf("request-%v.txt", etag)), c.Request)
	if writeError != nil {
		log.Println(writeError)
	}

	c.JSON(http.StatusOK, UploadDocumentResponseLocal(`http://bct-service.bct-service.svc.cluster.local:8080/api/graphql`))
}

func graphqlLocalHandler(c *gin.Context) {
	etag := requestid.Get(c)
	writeError := writeHttpRequestToFile(path.Join(dst, fmt.Sprintf("request-%v.txt", etag)), c.Request)
	if writeError != nil {
		log.Println(writeError)
	}

	c.JSON(http.StatusOK, UploadDocumentResponseLocal(`http://0.0.0.0:8080/api/graphql/local`))
}

func UploadDocumentResponseLocal(url string) *graphql.Response {
	return &graphql.Response{
		Errors: nil,
		Data: graphql.Data{
			DocumentQuery: graphql.DocumentQuery{
				GenerateUploadHyperlink: graphql.Hyperlink{
					Verb: http.MethodPost,
					URL:  url,
				},
			},
		},
	}
}

func writeHttpRequestToFile(filename string, req *http.Request) error {
	if req == nil {
		return nil
	}
	var bytes []byte

	// header
	var hdr = []byte("****header****")
	bytes = append(bytes, hdr...)
	for key, values := range req.Header {
		bytes = append(bytes, []byte(fmt.Sprintf("%s: %s", key, values[0]))...)
	}

	// body
	var bodyHdr = []byte("****body*****")
	bytes = append(bytes, bodyHdr...)

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	bytes = append(bytes, body...)

	return os.WriteFile(filename, bytes, 0644)
}
