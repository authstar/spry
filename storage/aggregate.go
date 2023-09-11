package storage

import (
	"context"
	"errors"
	"reflect"

	"github.com/authstar/spry"
	"github.com/gofrs/uuid"
)

type AggregateRepository[T spry.Aggregate[T]] struct {
	Repository[T]
}

func (repository AggregateRepository[T]) Fetch(ids spry.Identifiers) (T, error) {
	ctx := context.Background()
	ctx, err := repository.Storage.GetContext(ctx)
	if err != nil {
		return getEmpty[T](), err
	}

	identifiers := map[string][]spry.Identifiers{
		repository.ActorName: []map[string]any{ids},
	}
	assignments, err := repository.getAssignedIds(ctx, identifiers)
	if err != nil {
		return getEmpty[T](), err
	}
	snapshot, err := repository.fetchAggregate(ctx, assignments)
	if err != nil {
		return getEmpty[T](), err
	}
	return snapshot.Data.(T), nil
}

func (repository AggregateRepository[T]) Handle(command spry.Command) spry.Results[T] {
	ctx, err := repository.Storage.GetContext(context.Background())
	if err != nil {
		return spry.Results[T]{Errors: []error{err}}
	}
	if _, ok := command.(spry.Aggregate[T]); ok {
		return repository.handleAggregateCommand(ctx, command)
	}
	return spry.Results[T]{
		Errors: []error{errors.New("command must implement GetIdentifierSet")},
	}
}

func (repository AggregateRepository[T]) createSnapshot(next T, baseline Snapshot, cmdRecord CommandRecord, events []EventRecord) (Snapshot, spry.Results[T], bool) {
	lastEventRecord := events[len(events)-1]
	snapshot, err := NewSnapshot(next)
	if err != nil {
		return Snapshot{}, spry.Results[T]{
			Original: next,
			Errors:   []error{err},
		}, true
	}
	snapshot.ActorId = baseline.ActorId
	snapshot.LastCommandId = cmdRecord.Id
	snapshot.LastCommandOn = cmdRecord.HandledOn
	snapshot.LastEventId = lastEventRecord.Id
	snapshot.LastEventOn = lastEventRecord.CreatedOn
	snapshot.EventsApplied += uint64(len(events))
	snapshot.EventSinceSnapshot += len(events)
	snapshot.Version++

	for _, er := range events {
		if er.ActorName != repository.ActorName {
			snapshot.AddLastEventFor(er.ActorName, er.ActorId, er.Id)
		}
	}

	return snapshot, spry.Results[T]{}, false
}

func (repository Repository[T]) fetchAggregate(ctx context.Context, assignments IdAssignments) (Snapshot, error) {
	uid := assignments.GetAggregateId()

	// get the latest snapshot or initialize and empty
	snapshot, err := repository.getLatestSnapshotByUUID(ctx, uid)
	actorId := snapshot.ActorId
	if err != nil {
		return snapshot, err
	}

	// check for all events since the latest snapshot
	events, records, err := repository.getAggregatedEventsSince(ctx, actorId, snapshot)
	if err != nil {
		return snapshot, err
	}

	// apply events to actor instance
	repository.updateActor(events, records, &snapshot)

	// write snapshot
	err = repository.writeSnapshot(ctx, len(events), snapshot)
	return snapshot, err
}

func (repository Repository[T]) getAggregatedEventsSince(ctx context.Context, aggregateId uuid.UUID, snapshot Snapshot) ([]spry.Event, []EventRecord, error) {
	var err error

	// get id map from map store
	if aggregateId == uuid.Nil {
		return nil, nil, err
	}

	idMap, err := repository.Storage.FetchIdMap(ctx, repository.ActorName, aggregateId)
	if err != nil {
		return nil, nil, err
	}

	snapshot.UpdateFromMap(idMap)

	records, err := repository.Storage.FetchAggregatedEventsSince(
		ctx,
		snapshot.Type,
		snapshot.ActorId,
		snapshot.LastEventId,
		snapshot.LastEventMap,
	)
	if err != nil {
		return nil, nil, err
	}

	eventCount := len(records)
	events := make([]spry.Event, eventCount)
	for i, record := range records {
		events[i] = record.Data.(spry.Event)
	}
	return events, records, nil
}

func (repository AggregateRepository[T]) handleAggregateCommand(ctx context.Context, command spry.Command) spry.Results[T] {
	identifiers := command.(spry.Aggregate[T]).GetIdentifierSet()
	idSet := spry.IdSetFromIdentifierSet(identifiers)
	aggregateId := idSet.GetIdsFor(repository.ActorName)[0]

	assignments, err := repository.getAssignedIds(ctx, identifiers)
	if err != nil {
		return spry.Results[T]{
			Errors: []error{err},
		}
	}

	baseline, err := repository.fetchAggregate(ctx, assignments)
	if err != nil {
		return spry.Results[T]{
			Errors: []error{err},
		}
	}

	// fetch events for associated records

	cmdRecord, s, done := repository.createCommandRecord(command, baseline)
	if done {
		return s
	}

	actor := baseline.Data.(T)
	events, errors := command.Handle(actor)

	if len(errors) > 0 {
		return spry.Results[T]{
			Original: actor,
			Errors:   errors,
		}
	}

	next := repository.Apply(events, actor)
	eventRecords, s, done := repository.createEventRecords(events, baseline, cmdRecord, assignments)
	if done {
		return s
	}

	snapshot, s, done := repository.createSnapshot(next, baseline, cmdRecord, eventRecords)
	if done {
		return s
	}

	// store id map
	err = repository.Storage.AddMap(ctx, repository.ActorName, aggregateId, snapshot.ActorId)
	if err != nil {
		return spry.Results[T]{
			Original: actor,
			Modified: next,
			Events:   events,
			Errors:   []error{err},
		}
	}

	// store events
	err = repository.Storage.AddEvents(ctx, eventRecords)
	if err != nil {
		_ = repository.Storage.Rollback(ctx)
		return spry.Results[T]{
			Original: actor,
			Modified: next,
			Events:   events,
			Errors:   []error{err},
		}
	}

	config := spry.GetActorMeta[T]()
	// do we allow snapshotting during read?
	// if so, have we passed the event threshold?
	if config.SnapshotDuringWrite &&
		snapshot.EventSinceSnapshot >= config.SnapshotFrequency {
		snapshot.EventSinceSnapshot = 0
		err = repository.Storage.AddSnapshot(
			ctx,
			repository.ActorName,
			snapshot,
			config.SnapshotDuringPartition,
		)
		if err != nil {
			return spry.Results[T]{
				Original: actor,
				Modified: next,
				Events:   events,
				Errors:   []error{err},
			}
		}
	}

	err = repository.Storage.Commit(ctx)
	if err != nil {
		_ = repository.Storage.Rollback(ctx)
		return spry.Results[T]{
			Original: actor,
			Modified: next,
			Events:   events,
			Errors:   []error{err},
		}
	}

	return spry.Results[T]{
		Original: actor,
		Modified: next,
		Events:   events,
		Errors:   errors,
	}
}

func GetAggregateRepositoryFor[T spry.Aggregate[T]](storage Storage) AggregateRepository[T] {
	actorType := reflect.TypeOf(*new(T))
	actorName := actorType.Name()
	return AggregateRepository[T]{
		Repository: Repository[T]{
			ActorType: actorType,
			ActorName: actorName,
			Storage:   storage,
		},
	}
}
