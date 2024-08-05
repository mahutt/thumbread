package main

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mahutt/thumbread/pkg/request"
)

type Template struct {
	tmpl *template.Template
}

func newTemplate() *Template {
	return &Template{
		tmpl: template.Must(template.ParseGlob("web/views/*.html")),
	}
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.tmpl.ExecuteTemplate(w, name, data)
}

func handleIndex(c echo.Context) error {
	return c.Render(200, "index", nil)
}

func handleUrl(c echo.Context) error {
	url := c.Param("url")
	content, err := request.GetContent(url)
	if err != nil {
		return c.Render(500, "error", nil)
	}
	return c.Render(200, "content", map[string]interface{}{
		"content": content,
	})
}

func main() {
	e := echo.New()

	e.Static("/css", "web/css")
	e.File("/favicon.ico", "web/favicon.ico")

	e.Renderer = newTemplate()
	e.Use(middleware.Logger())

	e.GET("/", handleIndex)
	e.GET("/:url", handleUrl)

	e.Logger.Fatal(e.Start(":8080"))
}
