package main

import (
	"fmt"
	"runtime"

	"github.com/gin-gonic/gin"
)

var Router *gin.Engine

func main() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello world!",
		})
	})

	fmt.Printf("Hello from Go %s\n", runtime.Version())

	r.Run()
}
