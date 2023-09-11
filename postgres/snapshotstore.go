package postgres

import (
	"context"

	"github.com/authstar/spry"
	"github.com/authstar/spry/storage"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresSnapshotStore struct {
	Pool      *pgxpool.Pool
	Templates storage.StringTemplate
}

func (store *PostgresSnapshotStore) Add(ctx context.Context, actorName string, snapshot storage.Snapshot, allowPartition bool) error {
	query, _ := store.Templates.Execute(
		"insert_snapshot.sql",
		queryData(actorName),
	)
	tx := storage.GetTx[pgx.Tx](ctx)
	err := tx.BeginFunc(
		ctx,
		func(t pgx.Tx) error {
			data, err := spry.ToJson(snapshot)
			if err != nil {
				return err
			}
			_, err = t.Exec(
				ctx,
				query,
				snapshot.Id,
				snapshot.ActorId,
				data,
				snapshot.LastCommandId,
				snapshot.LastCommandOn,
				snapshot.LastEventId,
				snapshot.LastEventOn,
				snapshot.Version,
			)
			return err
		},
	)
	return err
}

func (store *PostgresSnapshotStore) Fetch(ctx context.Context, actorName string, actorId uuid.UUID) (storage.Snapshot, error) {
	query, _ := store.Templates.Execute(
		"select_latest_snapshot.sql",
		queryData(actorName),
	)
	tx := storage.GetTx[pgx.Tx](ctx)
	rows, err := tx.Query(
		ctx,
		query,
		actorId,
	)
	if err != nil {
		return storage.Snapshot{}, err
	}
	defer rows.Close()
	record := storage.Snapshot{}
	if rows.Next() {
		buffer := []byte{}
		err = rows.Scan(nil, &buffer, nil, nil, nil, nil, nil)
		if err != nil {
			return storage.Snapshot{}, err
		}
		record, err = spry.FromJson[storage.Snapshot](buffer)
		if err != nil {
			return record, err
		}
	}
	return record, nil
}
