package storage

import (
	"context"
	"errors"
	"reflect"

	"github.com/authstar/spry"
)

type ActorRepository[T spry.Actor[T]] struct {
	Repository[T]
}

func (repository ActorRepository[T]) Fetch(ids spry.Identifiers) (T, error) {
	ctx := context.Background()
	ctx, err := repository.Storage.GetContext(ctx)
	if err != nil {
		return getEmpty[T](), err
	}
	snapshot, err := repository.fetchActor(ctx, ids)
	if err != nil {
		return getEmpty[T](), err
	}
	return snapshot.Data.(T), nil
}

func (repository ActorRepository[T]) handleActorCommand(ctx context.Context, command spry.Command) spry.Results[T] {
	identifiers := command.(spry.Actor[T]).GetIdentifiers()
	baseline, err := repository.fetchActor(ctx, identifiers)
	if err != nil {
		return spry.Results[T]{
			Errors: []error{err},
		}
	}

	cmdRecord, s, done := repository.createCommandRecord(command, baseline)
	if done {
		return s
	}

	actor := baseline.Data.(T)
	events, errors := command.Handle(actor)
	next := repository.Apply(events, actor)
	eventRecords, s, done := repository.createEventRecords(events, baseline, cmdRecord, IdAssignments{})
	if done {
		return s
	}

	snapshot, s, done := repository.createSnapshot(next, baseline, cmdRecord, eventRecords)
	if done {
		return s
	}

	// store id map
	err = repository.Storage.AddMap(ctx, repository.ActorName, identifiers, snapshot.ActorId)
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

func (repository ActorRepository[T]) createSnapshot(next T, baseline Snapshot, cmdRecord CommandRecord, events []EventRecord) (Snapshot, spry.Results[T], bool) {
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
	return snapshot, spry.Results[T]{}, false
}

func (repository ActorRepository[T]) Handle(command spry.Command) spry.Results[T] {
	ctx, err := repository.Storage.GetContext(context.Background())
	if err != nil {
		return spry.Results[T]{Errors: []error{err}}
	}
	if _, ok := command.(spry.Actor[T]); ok {
		return repository.handleActorCommand(ctx, command)
	}
	return spry.Results[T]{
		Errors: []error{errors.New("command must implement GetIdentifiers")},
	}
}

func GetActorRepositoryFor[T spry.Actor[T]](storage Storage) ActorRepository[T] {
	actorType := reflect.TypeOf(*new(T))
	actorName := actorType.Name()
	return ActorRepository[T]{
		Repository: Repository[T]{
			ActorType: actorType,
			ActorName: actorName,
			Storage:   storage,
		},
	}
}
