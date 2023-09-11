package postgres

import (
	"embed"
	"fmt"
	"strings"
	"time"

	"github.com/authstar/spry/storage"
)

var banner = `
//******************************************************************************
//*                                                                            *
//*                    Generated PostgreSQL Schema
//*                    for %s
//*                    on %s
//*                                                                            *
//******************************************************************************
%s
`

//go:embed sql
var pgSqlFiles embed.FS

func PostgresGenerateActorSchema(actorName string) (string, error) {
	type Data struct {
		ActorName string
	}
	templates, err := storage.CreateTemplateFromFS(
		pgSqlFiles,
		"sql/create_actor_schema.sql",
	)
	if err != nil {
		return "", error(fmt.Errorf("failed to create templates from embedded FS: %e", err))
	}
	result, err := templates.Execute(
		"create_actor_schema.sql",
		Data{ActorName: strings.ToLower(actorName)},
	)
	if err != nil {
		return "", error(fmt.Errorf("failed to create schema from template: %e", err))
	}

	now := time.Now()
	stamp := now.Format("2006-01-02 15:04:05")
	return fmt.Sprintf(banner, actorName, stamp, result), nil
}
