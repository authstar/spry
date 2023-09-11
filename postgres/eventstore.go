package postgres

import (
	"context"
	"sort"

	"github.com/authstar/spry"
	"github.com/authstar/spry/storage"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresEventStore struct {
	Pool      *pgxpool.Pool
	Templates storage.StringTemplate
}

func (store *PostgresEventStore) Add(ctx context.Context, events []storage.EventRecord) error {
	tx := storage.GetTx[pgx.Tx](ctx)
	batch := pgx.Batch{}
	for _, event := range events {
		data, err := spry.ToJson(event)
		if err != nil {
			return err
		}
		query, _ := store.Templates.Execute(
			"insert_event.sql",
			queryData(event.ActorName),
		)
		batch.Queue(
			query,
			event.Id,
			event.ActorId,
			data,
			event.CreatedOn,
			event.CreatedByVersion,
		)
	}

	results := tx.SendBatch(ctx, &batch)
	_, _ = results.Exec()
	err := results.Close()
	return err
}

func (store *PostgresEventStore) FetchAggregatedSince(
	ctx context.Context,
	actorName string,
	actorId uuid.UUID,
	eventUUID uuid.UUID,
	idMap storage.LastEventMap,
	types storage.TypeMap) ([]storage.EventRecord, error) {

	var records []storage.EventRecord
	own, err := store.FetchSince(ctx, actorName, actorId, eventUUID, types)
	if err != nil {
		return nil, err
	}
	records = append(records, own...)

	for childName, childMap := range idMap.LastEvents {
		for id, last := range childMap {
			list, err := store.FetchSince(ctx, childName, id, last, types)
			if err != nil {
				return nil, err
			}
			records = append(records, list...)
		}
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].Id.String() < records[j].Id.String()
	})

	return records, nil
}

func (store *PostgresEventStore) FetchSince(
	ctx context.Context,
	actorName string,
	actorId uuid.UUID,
	eventUUID uuid.UUID,
	types storage.TypeMap) ([]storage.EventRecord, error) {
	query, _ := store.Templates.Execute(
		"select_events_since.sql",
		queryData(actorName),
	)
	tx := storage.GetTx[pgx.Tx](ctx)
	rows, err := tx.Query(
		ctx,
		query,
		actorId,
		eventUUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := []storage.EventRecord{}
	for rows.Next() {
		buffer := []byte{}
		err = rows.Scan(nil, nil, nil, &buffer, nil)
		if err != nil {
			return nil, err
		}
		record, err := spry.FromJson[storage.EventRecord](buffer)
		if err != nil {
			return nil, err
		}
		record.Data, err = types.AsEvent(record.Type, record.Data)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}
