package storage

import (
	"context"
	"reflect"

	"github.com/authstar/spry"
	"github.com/gofrs/uuid"
)

type NoOpTx struct{}

func (tx NoOpTx) Rollback() error {
	return nil
}
func (tx NoOpTx) Commit() error {
	return nil
}

var tx_key = reflect.TypeOf(NoOpTx{})

func GetTx[T any](ctx context.Context) T {
	return ctx.Value(tx_key).(T)
}

type CommandStore interface {
	Add(context.Context, string, CommandRecord) error
}

type EventStore interface {
	Add(context.Context, []EventRecord) error
	FetchAggregatedSince(context.Context, string, uuid.UUID, uuid.UUID, LastEventMap, TypeMap) ([]EventRecord, error)
	FetchSince(context.Context, string, uuid.UUID, uuid.UUID, TypeMap) ([]EventRecord, error)
}

type MapStore interface {
	AddId(context.Context, string, spry.Identifiers, uuid.UUID) error
	AddLink(context.Context, string, uuid.UUID, string, uuid.UUID) error
	GetId(context.Context, string, spry.Identifiers) (uuid.UUID, error)
	GetIdMap(context.Context, string, uuid.UUID) (AggregateIdMap, error)
}

type SnapshotStore interface {
	Add(context.Context, string, Snapshot, bool) error
	Fetch(context.Context, string, uuid.UUID) (Snapshot, error)
}

type TxProvider[T any] interface {
	GetTransaction(ctx context.Context) (T, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type Storage interface {
	AddCommand(context.Context, string, CommandRecord) error
	AddEvents(context.Context, []EventRecord) error
	AddMap(context.Context, string, spry.Identifiers, uuid.UUID) error
	AddSnapshot(context.Context, string, Snapshot, bool) error
	AddLink(context.Context, string, uuid.UUID, string, uuid.UUID) error
	Commit(context.Context) error
	FetchAggregatedEventsSince(context.Context, string, uuid.UUID, uuid.UUID, LastEventMap) ([]EventRecord, error)
	FetchEventsSince(context.Context, string, uuid.UUID, uuid.UUID) ([]EventRecord, error)
	FetchId(context.Context, string, spry.Identifiers) (uuid.UUID, error)
	FetchIdMap(context.Context, string, uuid.UUID) (AggregateIdMap, error)
	FetchLatestSnapshot(context.Context, string, uuid.UUID) (Snapshot, error)
	GetContext(context.Context) (context.Context, error)
	RegisterPrimitives(...any)
	Rollback(context.Context) error
}

type Stores[Tx any] struct {
	Commands     CommandStore
	Events       EventStore
	Maps         MapStore
	Primitives   TypeMap
	Snapshots    SnapshotStore
	Transactions TxProvider[Tx]
}

func (storage Stores[Tx]) AddCommand(ctx context.Context, actorName string, command CommandRecord) error {
	return storage.Commands.Add(ctx, actorName, command)
}

func (storage Stores[Tx]) AddEvents(ctx context.Context, events []EventRecord) error {
	return storage.Events.Add(ctx, events)
}

func (storage Stores[Tx]) AddLink(
	ctx context.Context,
	parentName string,
	parentId uuid.UUID,
	childName string,
	childId uuid.UUID) error {
	return storage.Maps.AddLink(ctx, parentName, parentId, childName, childId)
}

func (storage Stores[Tx]) AddMap(ctx context.Context, actorName string, identifiers spry.Identifiers, uid uuid.UUID) error {
	return storage.Maps.AddId(ctx, actorName, identifiers, uid)
}

func (storage Stores[Tx]) AddSnapshot(ctx context.Context, actorName string, snapshot Snapshot, allowPartition bool) error {
	return storage.Snapshots.Add(ctx, actorName, snapshot, allowPartition)
}

func (storage Stores[Tx]) Commit(ctx context.Context) error {
	return storage.Transactions.Commit(ctx)
}

func (storage Stores[Tx]) FetchAggregatedEventsSince(ctx context.Context, actorName string, actorId uuid.UUID, eventId uuid.UUID, idMap LastEventMap) ([]EventRecord, error) {
	return storage.Events.FetchAggregatedSince(ctx, actorName, actorId, eventId, idMap, storage.Primitives)
}

func (storage Stores[Tx]) FetchEventsSince(ctx context.Context, actorName string, actorId uuid.UUID, eventId uuid.UUID) ([]EventRecord, error) {
	return storage.Events.FetchSince(ctx, actorName, actorId, eventId, storage.Primitives)
}

func (storage Stores[Tx]) FetchId(ctx context.Context, actorName string, identifiers spry.Identifiers) (uuid.UUID, error) {
	return storage.Maps.GetId(ctx, actorName, identifiers)
}

func (storage Stores[Tx]) FetchIdMap(ctx context.Context, actorName string, actorId uuid.UUID) (AggregateIdMap, error) {
	return storage.Maps.GetIdMap(ctx, actorName, actorId)
}

func (storage Stores[Tx]) FetchLatestSnapshot(ctx context.Context, actorName string, actorId uuid.UUID) (Snapshot, error) {
	return storage.Snapshots.Fetch(ctx, actorName, actorId)
}

func (storage Stores[Tx]) GetContext(ctx context.Context) (context.Context, error) {
	newTx, err := storage.Transactions.GetTransaction(ctx)
	if err != nil {
		return nil, err
	}
	return context.WithValue(ctx, tx_key, newTx), nil
}

func (storage Stores[Tx]) RegisterPrimitives(types ...any) {
	storage.Primitives.AddTypes(types...)
}

func (storage Stores[Tx]) Rollback(ctx context.Context) error {
	return storage.Transactions.Rollback(ctx)
}

func NewStorage[Tx any](
	commands CommandStore,
	events EventStore,
	maps MapStore,
	snapshots SnapshotStore,
	txs TxProvider[Tx]) Storage {
	return Stores[Tx]{
		Events:       events,
		Commands:     commands,
		Maps:         maps,
		Snapshots:    snapshots,
		Transactions: txs,
		Primitives:   CreateTypeMap(),
	}
}
