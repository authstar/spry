package tests

import (
	"testing"

	"github.com/authstar/spry"
	"github.com/authstar/spry/storage"
	"github.com/gofrs/uuid"
)

func TestAggregateIds(t *testing.T) {

	ids := storage.CreateAggregateIdMap("parent", uuid.Nil)

	for i := 0; i < 3; i++ {
		id, _ := storage.GetId()
		ids.AddIdsFor("one", id)
	}

	for i := 0; i < 3; i++ {
		id, _ := storage.GetId()
		ids.AddIdsFor("two", id)
	}

	if len(ids.Aggregated["one"]) < 3 {
		t.Error("id set one should have had 3 ids")
	}

	if len(ids.Aggregated["two"]) < 3 {
		t.Error("id set one should have had 3 ids")
	}
}

func TestLastEventMapUpdate(t *testing.T) {
	a1, _ := storage.GetId()
	a1event, _ := storage.GetId()
	a2, _ := storage.GetId()
	a2event, _ := storage.GetId()
	a3, _ := storage.GetId()
	b1, _ := storage.GetId()
	b1event, _ := storage.GetId()
	b2, _ := storage.GetId()
	b2event, _ := storage.GetId()
	b3, _ := storage.GetId()

	p1, _ := storage.GetId()

	lastEvents := storage.CreateLastEvents()
	lastEvents.AddLastEventFor("A", a1, a1event)
	lastEvents.AddLastEventFor("A", a2, a2event)
	lastEvents.AddLastEventFor("B", b1, b1event)
	lastEvents.AddLastEventFor("B", b2, b2event)

	aggMap := storage.CreateAggregateIdMap("P", p1)
	aggMap.AddIdsFor("A", a1, a2, a3)
	aggMap.AddIdsFor("B", b1, b2, b3)

	lastEvents.UpdateFromMap(aggMap)

	if lastEvents.LastEvents["A"][a1] != a1event ||
		lastEvents.LastEvents["A"][a2] != a2event ||
		lastEvents.LastEvents["A"][a3] != uuid.Nil ||
		lastEvents.LastEvents["B"][b1] != b1event ||
		lastEvents.LastEvents["B"][b2] != b2event ||
		lastEvents.LastEvents["B"][b3] != uuid.Nil {
		t.Error("last event map did not contain the expected mappings")
	}
}

type temp struct {
	Id uuid.UUID
}

func (t temp) GetIdentifiers() spry.Identifiers {
	return spry.Identifiers{"id": t.Id}
}

func TestSnapshotMapUpdate(t *testing.T) {
	a1, _ := storage.GetId()
	a1event, _ := storage.GetId()
	a2, _ := storage.GetId()
	a2event, _ := storage.GetId()
	a3, _ := storage.GetId()
	b1, _ := storage.GetId()
	b1event, _ := storage.GetId()
	b2, _ := storage.GetId()
	b2event, _ := storage.GetId()
	b3, _ := storage.GetId()

	p1, _ := storage.GetId()

	tid, _ := storage.GetId()
	actor := temp{Id: tid}

	s, _ := storage.NewSnapshot(actor)
	s.AddLastEventFor("A", a1, a1event)
	s.AddLastEventFor("A", a2, a2event)
	s.AddLastEventFor("B", b1, b1event)
	s.AddLastEventFor("B", b2, b2event)

	aggMap := storage.CreateAggregateIdMap("P", p1)
	aggMap.AddIdsFor("A", a1, a2, a3)
	aggMap.AddIdsFor("B", b1, b2, b3)

	s.UpdateFromMap(aggMap)

	if s.LastEvents["A"][a1] != a1event ||
		s.LastEvents["A"][a2] != a2event ||
		s.LastEvents["A"][a3] != uuid.Nil ||
		s.LastEvents["B"][b1] != b1event ||
		s.LastEvents["B"][b2] != b2event ||
		s.LastEvents["B"][b3] != uuid.Nil {
		t.Error("last event map did not contain the expected mappings")
	}
}
