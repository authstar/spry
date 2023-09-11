package tests

import (
	"testing"

	"github.com/authstar/spry"
	"github.com/authstar/spry/postgres"
	"github.com/authstar/spry/storage"
	"github.com/authstar/spry/tests"
)

func TestHandleCommandSuccessfully(t *testing.T) {
	store := postgres.CreatePostgresStorage(CONNECTION_STRING)
	store.RegisterPrimitives(
		tests.PlayerCreated{},
		tests.PlayerDamaged{},
		tests.PlayerHealed{},
	)

	t.Cleanup(func() {
		_ = TruncateTables(
			"player_commands",
			"player_events",
			"player_id_map",
			"player_snapshots",
		)
	})

	repo := storage.GetActorRepositoryFor[tests.Player](store)
	results := repo.Handle(tests.CreatePlayer{Name: "Bob"})

	// create player
	expected := tests.PlayerCreated{Name: "Bob"}
	if len(results.Events) == 0 ||
		results.Events[0].(tests.PlayerCreated) != expected {
		t.Error("event was not generated or did not match expected output")
	}
	if results.Original.Name != "" {
		t.Error("original actor instance was modified but should not have been")
	}
	if results.Modified.Name != "Bob" {
		t.Error("modified actor did not contain expected state")
	}

	// damage player
	expected2 := tests.PlayerDamaged{Damage: 40}
	results2 := repo.Handle(tests.DamagePlayer{Name: "Bob", Damage: 40})
	if len(results2.Events) == 0 ||
		results2.Events[0].(tests.PlayerDamaged) != expected2 {
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
	expected3 := tests.PlayerHealed{Health: 10}
	results3 := repo.Handle(tests.HealPlayer{Name: "Bob", Health: 10})
	if len(results3.Events) == 0 ||
		results3.Events[0].(tests.PlayerHealed) != expected3 {
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
	store := postgres.CreatePostgresStorage(CONNECTION_STRING)
	store.RegisterPrimitives(
		tests.VehicleRegistered{},
	)
	motorists := storage.GetAggregateRepositoryFor[tests.Motorist](store)
	vehicles := storage.GetActorRepositoryFor[tests.Vehicle](store)

	t.Cleanup(func() {
		_ = TruncateTables(
			"vehicle_commands",
			"vehicle_events",
			"vehicle_id_map",
			"vehicle_links",
			"vehicle_snapshots",
			"motorist_commands",
			"motorist_events",
			"motorist_id_map",
			"motorist_links",
			"motorist_snapshots",
		)
	})

	m1id := tests.MotoristId{
		License: "008767890",
		State:   "CA",
	}

	v1id := tests.VehicleId{
		VIN: "001002003",
	}

	rv1 := tests.RegisterVehicle{
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
