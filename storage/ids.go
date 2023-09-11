package storage

import (
	"github.com/authstar/spry"
	"github.com/gofrs/uuid"
)

type AggregatedIds = map[string][]uuid.UUID

type AggregateIdMap struct {
	ActorName  string
	ActorId    uuid.UUID
	Aggregated AggregatedIds
}

func (idMap *AggregateIdMap) AddIdsFor(child string, id ...uuid.UUID) {
	ids := idMap.Aggregated
	if list, ok := ids[child]; ok {
		ids[child] = append(list, id...)
	} else {
		ids[child] = id
	}
}

func CreateAggregateIdMap(actorName string, actorId uuid.UUID) AggregateIdMap {
	return AggregateIdMap{
		ActorName:  actorName,
		ActorId:    actorId,
		Aggregated: AggregatedIds{},
	}
}

func EmptyAggregateIdMap() AggregateIdMap {
	return AggregateIdMap{}
}

type LastEventMap struct {
	LastEvents map[string]map[uuid.UUID]uuid.UUID
}

func (last *LastEventMap) AddLastEventFor(child string, childId uuid.UUID, lastEventId uuid.UUID) {
	events := last.LastEvents
	if m, ok := events[child]; ok {
		m[childId] = lastEventId
	} else {
		events[child] = map[uuid.UUID]uuid.UUID{}
		events[child][childId] = lastEventId
	}
}

func (last *LastEventMap) UpdateFromMap(idMap AggregateIdMap) {
	events := last.LastEvents
	for k, list := range idMap.Aggregated {
		if _, ok := events[k]; !ok {
			events[k] = map[uuid.UUID]uuid.UUID{}
			for _, id := range list {
				events[k][id] = uuid.Nil
			}
		} else {
			for _, id := range list {
				if _, ok := events[k][id]; !ok {
					events[k][id] = uuid.Nil
				}
			}
		}
	}
}

func CreateLastEvents() LastEventMap {
	return LastEventMap{
		LastEvents: map[string]map[uuid.UUID]uuid.UUID{},
	}
}

type IdAssignment struct {
	ActorName   string
	AssignedId  uuid.UUID
	Identifiers spry.Identifiers
	Json        string
}

func NewAssignment(name string, ids spry.Identifiers, id uuid.UUID) IdAssignment {
	json, _ := spry.IdentifiersToString(ids)
	return IdAssignment{
		ActorName:   name,
		AssignedId:  id,
		Identifiers: ids,
		Json:        json,
	}
}

type IdAssignments struct {
	aggregateName       string
	aggregateAssignment IdAssignment
	byIdentifier        map[string]IdAssignment
	byId                map[uuid.UUID]IdAssignment
}

func (a *IdAssignments) AddAssignment(name string, ids spry.Identifiers, id uuid.UUID) {
	n := NewAssignment(name, ids, id)
	a.byId[id] = n
	a.byIdentifier[n.Json] = n
	if n.ActorName == a.aggregateName {
		a.aggregateAssignment = n
	}
}

func (a *IdAssignments) CreateAssignment(name string, ids spry.Identifiers) uuid.UUID {
	id, _ := GetId()
	a.AddAssignment(name, ids, id)
	return id
}

func (a *IdAssignments) GetAggregateId() uuid.UUID {
	return a.aggregateAssignment.AssignedId
}

func (a *IdAssignments) GetIdFor(name string, ids spry.Identifiers) uuid.UUID {
	json, _ := spry.IdentifiersToString(ids)
	if id, ok := a.byIdentifier[json]; ok {
		return id.AssignedId
	} else {
		return uuid.Nil
	}
}

func NewAssignments(aggregateName string) IdAssignments {
	return IdAssignments{
		aggregateName: aggregateName,
		byIdentifier:  map[string]IdAssignment{},
		byId:          map[uuid.UUID]IdAssignment{},
	}
}
