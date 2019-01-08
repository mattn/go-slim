package main

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/labstack/echo"
	"github.com/mattn/go-slim"
)

func main() {
	t, err := slim.ParseFile("view/index.slim")
	if err != nil {
		log.Fatal(err)
	}

	var m sync.RWMutex
	values := []string{}

	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		m.RLock()
		defer m.RUnlock()
		c.Request().Header.Set("content-type", "text/html")
		return t.Execute(c.Response(), map[string]interface{}{
			"names": values,
		})
	})
	e.POST("/add", func(c echo.Context) error {
		name := strings.TrimSpace(c.Request().PostFormValue("name"))
		if name != "" {
			m.Lock()
			defer m.Unlock()
			values = append(values, name)
		}
		return c.Redirect(http.StatusFound, "/")
	})
	e.Start(":8081")
}
