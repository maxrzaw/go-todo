package template

import (
	"errors"
	"io"
	"text/template"

	"github.com/labstack/echo/v4"
)

type TemplateRecipe struct {
	Name  string
	Base  string
	Paths []string
}

type TemplateRegistry struct {
	templates map[string]*template.Template
	bases     map[string]string
}

func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.templates[name]
	if !ok {
		err := errors.New("Template not found: " + name)
		return err
	}
	base, ok := t.bases[name]
	if !ok {
		err := errors.New("Template base not found: " + name)
		return err
	}
	return tmpl.ExecuteTemplate(w, base, data)
}

func NewTemplateRenderer(e *echo.Echo, recipes ...TemplateRecipe) {
	templs := make(map[string]*template.Template)
	bases := make(map[string]string)
	for i := range recipes {
		templs[recipes[i].Name] = template.Must(template.ParseFiles(recipes[i].Paths...))
		bases[recipes[i].Name] = recipes[i].Base
	}
	e.Renderer = &TemplateRegistry{
		templates: templs,
		bases:     bases,
	}
}
