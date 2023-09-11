package memory

import (
	"context"
	"sort"

	"github.com/authstar/spry"
	"github.com/authstar/spry/storage"
	"github.com/gofrs/uuid"
)

type IdLinks map[string]map[uuid.UUID]storage.AggregatedIds

type InMemoryCommandStore struct {
	Commands map[uuid.UUID][]storage.CommandRecord
}

func GetEventsAfter(events []storage.EventRecord, last uuid.UUID) []storage.EventRecord {
	after := []storage.EventRecord{}
	for _, e := range events {
		if e.Id.String() > last.String() {
			after = append(after, e)
		}
	}
	return after
}

func (store *InMemoryCommandStore) Add(
	ctx context.Context,
	actorName string,
	command storage.CommandRecord) error {
	if store.Commands == nil {
		store.Commands = map[uuid.UUID][]storage.CommandRecord{}
	}
	actorId := command.HandledBy
	if stored, ok := store.Commands[actorId]; ok {
		store.Commands[actorId] = append(stored, command)
	} else {
		store.Commands[actorId] = []storage.CommandRecord{command}
	}
	return nil
}

type InMemoryEventStore struct {
	Events map[uuid.UUID][]storage.EventRecord
}

func (store *InMemoryEventStore) Add(ctx context.Context, events []storage.EventRecord) error {
	if store.Events == nil {
		store.Events = map[uuid.UUID][]storage.EventRecord{}
	}
	for _, event := range events {
		actorId := event.ActorId
		if stored, ok := store.Events[actorId]; ok {
			store.Events[actorId] = append(stored, event)
		} else {
			store.Events[actorId] = []storage.EventRecord{event}
		}
	}
	return nil
}

func (store *InMemoryEventStore) FetchAggregatedSince(
	ctx context.Context,
	actorName string,
	actorId uuid.UUID,
	eventUUID uuid.UUID,
	idMap storage.LastEventMap,
	types storage.TypeMap) ([]storage.EventRecord, error) {

	if store.Events == nil {
		store.Events = map[uuid.UUID][]storage.EventRecord{}
	}

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

func (store *InMemoryEventStore) FetchSince(
	ctx context.Context,
	actorName string,
	actorId uuid.UUID,
	eventUUID uuid.UUID,
	types storage.TypeMap) ([]storage.EventRecord, error) {
	if store.Events == nil {
		store.Events = map[uuid.UUID][]storage.EventRecord{}
	}
	if stored, ok := store.Events[actorId]; ok {
		return GetEventsAfter(stored, eventUUID), nil
	} else {
		return []storage.EventRecord{}, nil
	}
}

type InMemoryMapStore struct {
	IdMap   map[string]uuid.UUID
	LinkMap IdLinks
}

func (maps *InMemoryMapStore) AddId(ctx context.Context, actorName string, ids spry.Identifiers, uid uuid.UUID) error {
	if maps.IdMap == nil {
		maps.IdMap = map[string]uuid.UUID{}
	}
	key, _ := spry.IdentifiersToString(ids)
	maps.IdMap[key] = uid
	return nil
}

func (maps *InMemoryMapStore) AddLink(ctx context.Context, parentType string, parentId uuid.UUID, childType string, childId uuid.UUID) error {
	if maps.LinkMap == nil {
		maps.LinkMap = IdLinks{}
	}
	if maps.LinkMap[parentType] == nil {
		maps.LinkMap[parentType] = map[uuid.UUID]storage.AggregatedIds{}
	}
	if maps.LinkMap[parentType][parentId] == nil {
		maps.LinkMap[parentType][parentId] = storage.AggregatedIds{}
	}
	if maps.LinkMap[parentType][parentId][childType] == nil {
		maps.LinkMap[parentType][parentId][childType] = []uuid.UUID{childId}
	} else {
		maps.LinkMap[parentType][parentId][childType] = append(maps.LinkMap[parentType][parentId][childType], childId)
	}
	return nil
}

func (maps *InMemoryMapStore) GetId(ctx context.Context, actorName string, ids spry.Identifiers) (uuid.UUID, error) {
	if maps.IdMap == nil {
		maps.IdMap = map[string]uuid.UUID{}
	}
	key, _ := spry.IdentifiersToString(ids)
	uid := maps.IdMap[key]
	if uid == uuid.Nil {
		return uuid.Nil, nil
	}
	return uid, nil
}

func (maps *InMemoryMapStore) GetIdMap(
	ctx context.Context,
	actorName string,
	uid uuid.UUID) (storage.AggregateIdMap, error) {
	if maps.IdMap == nil {
		maps.IdMap = map[string]uuid.UUID{}
	}

	ok := false
	var aggregates map[uuid.UUID]storage.AggregatedIds
	if aggregates, ok = maps.LinkMap[actorName]; !ok {
		aggregates = map[uuid.UUID]storage.AggregatedIds{}
		maps.LinkMap[actorName] = aggregates
	}

	idMap := storage.CreateAggregateIdMap(actorName, uid)
	if actors, ok := aggregates[uid]; ok {
		for k, v := range actors {
			idMap.AddIdsFor(k, v...)
		}
	}

	return idMap, nil
}

type InMemorySnapshotStore struct {
	Snapshots map[uuid.UUID][]storage.Snapshot
}

func (store *InMemorySnapshotStore) Add(ctx context.Context, actorName string, snapshot storage.Snapshot, allowPartition bool) error {
	if store.Snapshots == nil {
		store.Snapshots = map[uuid.UUID][]storage.Snapshot{}
	}
	actorId := snapshot.ActorId
	if stored, ok := store.Snapshots[actorId]; ok {
		store.Snapshots[actorId] = append(stored, snapshot)
	} else {
		store.Snapshots[actorId] = []storage.Snapshot{snapshot}
	}
	return nil
}

func (store *InMemorySnapshotStore) Fetch(ctx context.Context, actorName string, actorId uuid.UUID) (storage.Snapshot, error) {
	if store.Snapshots == nil {
		store.Snapshots = map[uuid.UUID][]storage.Snapshot{}
	}
	if stored, ok := store.Snapshots[actorId]; ok {
		return stored[len(stored)-1], nil
	}
	return storage.Snapshot{}, nil
}

type InMemoryTxProvider struct {
}

func (provider InMemoryTxProvider) Commit(ctx context.Context) error {
	return nil
}

func (provider InMemoryTxProvider) GetTransaction(ctx context.Context) (storage.NoOpTx, error) {
	return storage.NoOpTx{}, nil
}

func (provider InMemoryTxProvider) Rollback(ctx context.Context) error {
	return nil
}

func InMemoryStorage() storage.Storage {
	return storage.NewStorage[storage.NoOpTx](
		&InMemoryCommandStore{},
		&InMemoryEventStore{},
		&InMemoryMapStore{
			IdMap:   map[string]uuid.UUID{},
			LinkMap: IdLinks{},
		},
		&InMemorySnapshotStore{},
		&InMemoryTxProvider{},
	)
}
