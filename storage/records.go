package storage

import (
	"reflect"
	"time"

	"github.com/authstar/spry"
	"github.com/gofrs/uuid"
)

func GetId() (uuid.UUID, error) {
	UUID, err := uuid.NewV6()
	if err != nil {
		return uuid.Nil, err
	}
	return UUID, nil
}

type Snapshot struct {
	// aggregates track records and last events from each
	// for each child type, track the last event per Identifier
	LastEventMap
	// a generated uuid (system id) for the snapshot instance
	Id uuid.UUID `json:"id"`
	// this is the addressable identity of the owning model
	ActorId uuid.UUID `json:"actorId"`
	// the type name of the actor
	Type string `json:"type"`
	// a serialized causal tracker
	Vector string `json:"vector"`
	// a numeric version of the model
	Version uint64 `json:"version"`
	// the causal tracker of the preceding snapshot
	Ancestor string `json:"ancestor"`
	// UTC ISO date time string when event was created
	CreatedOn time.Time `json:"createdOn"`
	// the number of events applied to reach the present state
	EventsApplied uint64 `json:"eventsApplied"`
	// the number of events since the last snapshot was created
	EventSinceSnapshot int
	// the UUID of the last event played against the instance
	LastEventId uuid.UUID `json:"lastEventId"`
	// the UUID of the last command handled
	LastCommandId uuid.UUID `json:"lastCommandId"`
	// the wall clock at the time of the last command
	LastCommandOn time.Time `json:"lastCommandOn"`
	// the wall clock at the time of the last event
	LastEventOn time.Time `json:"lastEventOn"`
	// the contents of the snapshot
	Data any `json:"data"`
}

func (snapshot Snapshot) IsValid() bool {
	return !snapshot.Id.IsNil()
}

func NewSnapshot(actor any) (Snapshot, error) {
	actorType := reflect.TypeOf(actor)
	actorName := actorType.Name()
	id, err := GetId()
	if err != nil {
		return Snapshot{}, err
	}

	return Snapshot{
		Id:        id,
		Type:      actorName,
		CreatedOn: time.Now().UTC(),
		Data:      actor,
		LastEventMap: LastEventMap{
			LastEvents: map[string]map[uuid.UUID]uuid.UUID{},
		},
	}, nil
}

type EventRecord struct {
	// a generated uuid for this event
	Id uuid.UUID `json:"id"`
	// the type name of the event
	Type string `json:"type"`
	// inferred from the actor emitting the event
	ActorNamespace string `json:"namespace"`
	// this is the addressable identity of the owning model
	ActorId uuid.UUID `json:"actorId"`
	// the type of the model the event was generated for
	ActorName string `json:"actor"`
	// UTC ISO date time string when event was created
	CreatedOn time.Time `json:"createdOn"`
	// the type of the actor instantiating the event
	CreatedBy string `json:"createdBy"`
	// the id of the snapshot instantiating the event
	CreatedById uuid.UUID `json:"createdById"`
	// the vector of the snapshot instantiating the event
	CreatedByVector string `json:"createdByVector"`
	// the version of the snapshot instantiating the event
	CreatedByVersion uint64 `json:"createdByVersion"`
	// the command type/topic that triggered the event
	InitiatedBy string `json:"initiatedBy"`
	// the id of the message that triggered the event
	InitiatedById uuid.UUID `json:"initiatedById"`
	// the contents of the event
	Data any `json:"data"`
}

func (event EventRecord) IsValid() bool {
	return event.Id.IsNil()
}

func NewEventRecord(event spry.Event) (EventRecord, error) {
	eventType := reflect.TypeOf(event)
	eventName := eventType.Name()
	id, err := GetId()
	if err != nil {
		return EventRecord{}, err
	}
	er := EventRecord{
		Id:        id,
		Type:      eventName,
		CreatedOn: time.Now().UTC(),
		Data:      event,
	}
	if n, ok := event.(spry.Namespaced); ok {
		meta := n.GetEventMeta()
		er.ActorName = meta.CreatedFor
		er.CreatedBy = meta.CreatedBy
	}
	return er, nil
}

type CommandRecord struct {
	// a generated uuid for this event
	Id uuid.UUID `json:"id"`
	// the type name of the command
	Type string `json:"type"`
	// namespace for the command
	Namespace string `json:"namespace"`
	// the time the command was handled
	CreatedOn time.Time `json:"createdOn"`
	// the time the command was handled
	ReceivedOn time.Time `json:"receivedOn"`
	// the time the command was handled
	HandledOn time.Time `json:"handledOn"`
	// the id of the recipient actor
	HandledBy uuid.UUID `json:"handledBy"`
	// the version of the actor that handled the command
	HandledVersion uint64
	// the contents of the command
	Data any `json:"data"`
}

func (command CommandRecord) IsValid() bool {
	return command.Id.IsNil()
}

func NewCommandRecord(command spry.Command) (CommandRecord, error) {
	commandType := reflect.TypeOf(command)
	commandName := commandType.Name()
	id, err := GetId()
	if err != nil {
		return CommandRecord{}, err
	}

	return CommandRecord{
		Id:        id,
		Type:      commandName,
		CreatedOn: time.Now().UTC(),
		Data:      command,
	}, nil
}
