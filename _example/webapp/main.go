package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-slim"
)

func main() {
	t, err := slim.ParseFile("view/index.slim")
	if err != nil {
		log.Fatal(err)
	}

	values := []string{}

	gin.DefaultWriter = colorable.NewColorableStderr()
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		t.Execute(c.Writer, map[string]interface{}{
			"names": values,
		})
	})
	r.POST("/add", func(c *gin.Context) {
		name := strings.TrimSpace(c.PostForm("name"))
		if name != "" {
			values = append(values, name)
		}
		c.Redirect(http.StatusFound, "/")
	})
	r.Run(":8081")
}
