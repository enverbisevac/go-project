package assets

import (
	"embed"
	"io/fs"
	"text/template"
)

const (
	layoutsDir   = "templates/layouts"
	templatesDir = "templates"
	extension    = "/*.html"
)

var (
	//go:embed templates/* templates/layouts/*
	Files embed.FS

	//go:embed migration
	MigrationFS embed.FS

	Templates map[string]*template.Template
)

func init() {
	err := LoadTemplates()
	if err != nil {
		panic(err)
	}
}

func LoadTemplates() error {
	if Templates == nil {
		Templates = make(map[string]*template.Template)
	}
	tmplFiles, err := fs.ReadDir(Files, templatesDir)
	if err != nil {
		return err
	}
	for _, tmpl := range tmplFiles {
		if tmpl.IsDir() {
			continue
		}

		pt, err := template.ParseFS(Files, templatesDir+"/"+tmpl.Name(), layoutsDir+extension)
		if err != nil {
			return err
		}

		Templates[tmpl.Name()] = pt
	}
	return nil
}
