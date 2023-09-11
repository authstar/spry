package tests

import (
	"embed"
	"strings"
	"testing"

	"github.com/authstar/spry/storage"
)

//go:embed sql
var sqlFiles embed.FS

func TestCreateTemplate(t *testing.T) {
	type Data struct {
		ActorName string
	}
	templates, err := storage.CreateTemplateFromFS(
		sqlFiles,
		"sql/create_actor_schema.sql",
	)
	if err != nil {
		t.Error("failed to create templates from embedded FS", err)
	}
	result, err := templates.Execute(
		"create_actor_schema.sql",
		Data{ActorName: strings.ToLower("TesT")},
	)
	if err != nil {
		t.Error("failed to create schema from template", err)
	}
	if result == "" {
		t.Error(result)
	}
}
