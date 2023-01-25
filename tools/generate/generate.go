package main

import (
	"embed"
	"encoding/json"
	"log"
	"os"
	"text/template"

	"github.com/FreekingDean/proxmox-api-go/pkg/jsonschema"
)

//go:embed templates
var templates embed.FS

func main() {
	tmpl, err := template.ParseFS(templates, "templates/*.go.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	d, err := os.ReadFile("../proxmox-api-go/scripts/schema.json")
	if err != nil {
		log.Fatal(err)
	}
	schemas, err := parseJSON(d)
	if err != nil {
		log.Fatal(err)
	}
	loadSchema(schemas, tmpl)
}

func loadSchema(s []*jsonschema.Schema, tmpl *template.Template) {
	for _, schema := range s {
		if source, ok := sources[schema.Path]; ok {
			if source.Get {
				source.Schema = schema
				source.ParseSource(schema.Info["GET"])
				err := tmpl.ExecuteTemplate(os.Stdout, "data_source", source)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
		if len(schema.Children) > 0 {
			loadSchema(schema.Children, tmpl)
		}
	}
}

func parseJSON(data []byte) ([]*jsonschema.Schema, error) {
	out := make([]*jsonschema.Schema, 0)
	err := json.Unmarshal(data, &out)
	return out, err
}
