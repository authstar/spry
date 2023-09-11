package tests

import (
	"testing"

	"github.com/authstar/spry"

	"github.com/authstar/spry/memory"
	"github.com/authstar/spry/storage"
)

func TestGetRepositoryFor(t *testing.T) {
	store := memory.InMemoryStorage()
	repo := storage.GetActorRepositoryFor[Player](store)
	if repo.ActorName != "Player" {
		t.Error("actor name was not the expected type")
	}
}

func TestActorHandlesCommandSuccessfully(t *testing.T) {
	store := memory.InMemoryStorage()
	repo := storage.GetActorRepositoryFor[Player](store)
	results := repo.Handle(CreatePlayer{Name: "Bob"})

	// create player
	expected := PlayerCreated{Name: "Bob"}
	if len(results.Events) == 0 ||
		results.Events[0].(PlayerCreated) != expected {
		t.Error("event was not generated or did not match expected output")
	}
	if results.Original.Name != "" {
		t.Error("original actor instance was modified but should not have been")
	}
	if results.Modified.Name != "Bob" {
		t.Error("modified actor did not contain expected state")
	}

	// damage player
	expected2 := PlayerDamaged{Damage: 40}
	results2 := repo.Handle(DamagePlayer{Name: "Bob", Damage: 40})
	if len(results2.Events) == 0 ||
		results2.Events[0].(PlayerDamaged) != expected2 {
		t.Error("event was not generated or did not match expected output")
	}
	if results2.Original.HitPoints != 100 {
		t.Error("original actor instance was modified but should not have been")
	}
	if results2.Modified.HitPoints != 60 {
		t.Error("modified actor did not contain expected state")
	}
	if results2.Modified.Name != "Bob" {
		t.Error("incorrect creation of new actor occurred")
	}

	// heal player
	expected3 := PlayerHealed{Health: 10}
	results3 := repo.Handle(HealPlayer{Name: "Bob", Health: 10})
	if len(results3.Events) == 0 ||
		results3.Events[0].(PlayerHealed) != expected3 {
		t.Error("event was not generated or did not match expected output")
	}
	if results3.Original.HitPoints != 60 {
		t.Error("original actor instance was modified but should not have been")
	}
	if results3.Modified.HitPoints != 70 {
		t.Errorf("expected player health to = %d but was %d", 70, results.Modified.HitPoints)
	}
	if results3.Modified.Name != "Bob" {
		t.Error("incorrect creation of new actor occurred")
	}
}

func TestAggregateHandlesCommandSuccessfully(t *testing.T) {
	store := memory.InMemoryStorage()
	motorists := storage.GetAggregateRepositoryFor[Motorist](store)
	vehicles := storage.GetActorRepositoryFor[Vehicle](store)

	m1id := MotoristId{
		License: "008767890",
		State:   "CA",
	}

	v1id := VehicleId{
		VIN: "001002003",
	}

	rv1 := RegisterVehicle{
		MotoristId: m1id,
		VehicleId:  v1id,
		Type:       "Moped",
		Make:       "Hyundai",
		Model:      "Scootchum",
		Color:      "Blurple",
	}
	r1 := motorists.Handle(rv1)
	if len(r1.Modified.Vehicles) < 1 {
		t.Error("expected motorist to have 1 vehicle after registration")
	}

	v1, _ := vehicles.Fetch(spry.Identifiers{"VIN": v1id.VIN})
	if v1.VIN != v1id.VIN {
		t.Error("failed to retain VIN")
	}

	m1, _ := motorists.Fetch(spry.Identifiers{"License": "008767890", "State": "CA"})
	mv1 := m1.Vehicles[0]
	if m1.License != m1id.License ||
		m1.State != m1id.State ||
		len(m1.Vehicles) != 1 ||
		mv1.Color != rv1.Color ||
		mv1.Make != rv1.Make ||
		mv1.Model != rv1.Model ||
		mv1.Type != rv1.Type ||
		mv1.VIN != rv1.VIN {
		t.Error("failed to rehydrate motorist correctly")
	}
}
