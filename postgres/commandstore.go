package postgres

import (
	"context"

	"github.com/authstar/spry"
	"github.com/authstar/spry/storage"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresCommandStore struct {
	Pool      *pgxpool.Pool
	Templates storage.StringTemplate
}

func (store *PostgresCommandStore) Add(ctx context.Context, actorName string, command storage.CommandRecord) error {
	query, _ := store.Templates.Execute(
		"insert_command.sql",
		queryData(actorName),
	)

	tx := storage.GetTx[pgx.Tx](ctx)

	err := tx.BeginFunc(
		ctx,
		func(t pgx.Tx) error {
			data, err := spry.ToJson(command)
			if err != nil {
				return err
			}
			_, err = t.Exec(
				ctx,
				query,
				command.Id,
				command.HandledBy,
				data,
				command.CreatedOn,
				command.HandledVersion,
			)
			return err
		},
	)
	return err
}
