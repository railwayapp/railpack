package main

import (
	"fmt"
	"runtime"

	greeting "gin"
	"gin/cmd/common"

	"github.com/gin-gonic/gin"
)

var Router *gin.Engine

func main() {
	banner := common.FormatBanner(greeting.Hello(runtime.Version()))

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": banner,
		})
	})

	fmt.Println(banner)

	r.Run()
}
