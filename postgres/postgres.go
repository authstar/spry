package postgres

import (
	"context"
	"embed"
	"fmt"

	"github.com/authstar/spry/storage"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type QueryData struct {
	ActorName string
}

//go:embed sql
var sqlFiles embed.FS

func queryData(name string) QueryData {
	return QueryData{ActorName: name}
}

func CreatePostgresStorage(connectionURI string) storage.Storage {
	pool, err := pgxpool.Connect(context.Background(), connectionURI)
	if err != nil {
		fmt.Println("failed to connect to the backing store", err)
		panic("oh no")
	}

	// load templates
	templates, err := storage.CreateTemplateFromFS(
		sqlFiles,
		"sql/insert_command.sql",
		"sql/insert_event.sql",
		"sql/insert_link.sql",
		"sql/insert_map.sql",
		"sql/insert_snapshot.sql",
		"sql/select_events_since.sql",
		"sql/select_id_by_map.sql",
		"sql/select_latest_snapshot.sql",
		"sql/select_links_for_actor.sql",
	)

	if err != nil {
		fmt.Println("failed to read sql templates")
		panic("oh no")
	}

	return storage.NewStorage[pgx.Tx](
		&PostgresCommandStore{Templates: *templates, Pool: pool},
		&PostgresEventStore{Templates: *templates, Pool: pool},
		&PostgresMapStore{Templates: *templates, Pool: pool},
		&PostgresSnapshotStore{Templates: *templates, Pool: pool},
		&PostgresTxProvider{Pool: pool},
	)
}
