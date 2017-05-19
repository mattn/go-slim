package main

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-slim"
)

func main() {
	t, err := slim.ParseFile("view/index.slim")
	if err != nil {
		log.Fatal(err)
	}

	var m sync.RWMutex
	values := []string{}

	gin.DefaultWriter = colorable.NewColorableStderr()
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		m.RLock()
		defer m.RUnlock()
		c.Header("content-type", "text/html")
		t.Execute(c.Writer, map[string]interface{}{
			"names": values,
		})
	})
	r.POST("/add", func(c *gin.Context) {
		name := strings.TrimSpace(c.PostForm("name"))
		if name != "" {
			m.Lock()
			defer m.Unlock()
			values = append(values, name)
		}
		c.Redirect(http.StatusFound, "/")
	})
	r.Run(":8081")
}
