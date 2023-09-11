package postgres

import (
	"context"
	"time"

	"github.com/authstar/spry"
	"github.com/authstar/spry/storage"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresMapStore struct {
	Pool      *pgxpool.Pool
	Templates storage.StringTemplate
}

func (store *PostgresMapStore) AddId(ctx context.Context, actorName string, ids spry.Identifiers, uid uuid.UUID) error {
	query, _ := store.Templates.Execute(
		"insert_map.sql",
		queryData(actorName),
	)
	tx := storage.GetTx[pgx.Tx](ctx)
	id, _ := storage.GetId()
	err := tx.BeginFunc(
		ctx,
		func(t pgx.Tx) error {
			data, err := spry.ToJson(ids)
			if err != nil {
				return err
			}
			_, err = t.Exec(
				ctx,
				query,
				id,
				data,
				uid,
				time.Now(),
			)
			return err
		},
	)
	return err
}

func (store *PostgresMapStore) AddLink(ctx context.Context, parentType string, parentId uuid.UUID, childType string, childId uuid.UUID) error {
	query, _ := store.Templates.Execute(
		"insert_link.sql",
		queryData(parentType),
	)
	tx := storage.GetTx[pgx.Tx](ctx)
	id, _ := storage.GetId()
	err := tx.BeginFunc(
		ctx,
		func(t pgx.Tx) error {
			_, err := t.Exec(
				ctx,
				query,
				id,
				parentType,
				parentId,
				childType,
				childId,
			)
			return err
		},
	)
	return err
}

func (store *PostgresMapStore) GetId(ctx context.Context, actorName string, ids spry.Identifiers) (uuid.UUID, error) {
	query, _ := store.Templates.Execute(
		"select_id_by_map.sql",
		queryData(actorName),
	)
	data, err := spry.ToJson(ids)
	if err != nil {
		return uuid.Nil, err
	}

	tx := storage.GetTx[pgx.Tx](ctx)
	rows, err := tx.Query(
		ctx,
		query,
		data,
	)
	if err != nil {
		return uuid.Nil, err
	}
	defer rows.Close()
	uid := uuid.Nil
	if rows.Next() {
		err = rows.Scan(nil, nil, &uid, nil)
		if err != nil {
			return uid, err
		}
	}
	return uid, nil
}

func (store *PostgresMapStore) GetIdMap(ctx context.Context, actorName string, uid uuid.UUID) (storage.AggregateIdMap, error) {
	query, _ := store.Templates.Execute(
		"select_links_for_actor.sql",
		queryData(actorName),
	)

	empty := storage.EmptyAggregateIdMap()
	tx := storage.GetTx[pgx.Tx](ctx)
	rows, err := tx.Query(
		ctx,
		query,
		actorName,
		uid,
	)
	if err != nil {
		return empty, err
	}
	defer rows.Close()

	idMap := storage.CreateAggregateIdMap(actorName, uid)

	for rows.Next() {
		var child string
		var id uuid.UUID
		err = rows.Scan(nil, nil, &child, &id)
		if err != nil {
			return empty, err
		}
		idMap.AddIdsFor(child, id)
	}

	return idMap, nil
}
