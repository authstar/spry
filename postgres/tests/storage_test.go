package tests

import (
	"context"
	"testing"
	"time"

	"github.com/authstar/spry"
	"github.com/authstar/spry/postgres"
	"github.com/authstar/spry/storage"
	"github.com/authstar/spry/tests"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
)

func TestCommandStorage(t *testing.T) {
	store := postgres.CreatePostgresStorage(
		CONNECTION_STRING,
	)

	uid, _ := storage.GetId()
	c1 := tests.CreatePlayer{Name: "Bob"}
	cr1, _ := storage.NewCommandRecord(c1)

	cr1.CreatedOn = time.Now()
	cr1.Data = c1
	cr1.HandledBy = uid
	cr1.HandledOn = time.Now()
	cr1.HandledVersion = 0
	cr1.Id = uid
	cr1.ReceivedOn = time.Now()
	cr1.Type = "CreatePlayer"

	ctx, _ := store.GetContext(context.Background())

	err := store.AddCommand(ctx, "Player", cr1)
	if err != nil {
		t.Fatal("failed to store command correctly", err)
	}

	tx := storage.GetTx[pgx.Tx](ctx)
	err = tx.Rollback(ctx)
	if err != nil {
		t.Error(err)
	}
}

func TestActorEventStorage(t *testing.T) {
	store := postgres.CreatePostgresStorage(
		CONNECTION_STRING,
	)
	store.RegisterPrimitives(
		tests.PlayerCreated{},
		tests.PlayerDamaged{},
		tests.PlayerHealed{},
	)

	aid1, _ := storage.GetId()

	e1 := tests.PlayerCreated{Name: "Bill"}
	er1, _ := storage.NewEventRecord(e1)
	er1.ActorId = aid1
	er1.ActorName = "Player"
	er1.CreatedBy = "Player"
	er1.CreatedById = aid1
	er1.Id, _ = storage.GetId()
	er1.Data = e1
	er1.Type = "PlayerCreated"

	e2 := tests.PlayerDamaged{Damage: 10}
	er2, _ := storage.NewEventRecord(e2)
	er2.ActorId = aid1
	er2.ActorName = "Player"
	er2.CreatedBy = "Player"
	er2.CreatedById = aid1
	er2.Id, _ = storage.GetId()
	er2.Data = e1
	er2.Type = "PlayerCreated"

	ctx, _ := store.GetContext(context.Background())

	err := store.AddEvents(ctx, []storage.EventRecord{
		er1,
		er2,
	})

	if err != nil {
		t.Error(err)
	}

	records, err := store.FetchEventsSince(ctx, "player", aid1, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("expected %d records but got %d instead", 2, len(records))
	}

	tx := storage.GetTx[pgx.Tx](ctx)
	err = tx.Rollback(ctx)
	if err != nil {
		t.Error(err)
	}
}

func TestAggregateEventStorage(t *testing.T) {
	store := postgres.CreatePostgresStorage(
		CONNECTION_STRING,
	)
	store.RegisterPrimitives(
		tests.VehicleRegistered{},
	)

	aid1, _ := storage.GetId()
	agid, _ := storage.GetId()

	e1 := tests.VehicleRegistered{
		MotoristId: tests.MotoristId{License: "123", State: "AK"},
		VehicleId:  tests.VehicleId{VIN: "000000001"},
		Type:       "scooter",
		Make:       "gogeddums",
		Model:      "scootchies",
		Color:      "arglebargle",
	}

	er1, _ := storage.NewEventRecord(e1)
	er1.ActorId = aid1
	er1.ActorName = "Vehicle"
	er1.CreatedBy = "Motorist"
	er1.CreatedById = agid
	er1.Id, _ = storage.GetId()
	er1.Data = e1
	er1.Type = "VehicleRegistered"

	ctx, _ := store.GetContext(context.Background())

	err := store.AddEvents(ctx, []storage.EventRecord{er1})
	if err != nil {
		t.Error(err)
	}

	records, err := store.FetchEventsSince(ctx, "vehicle", aid1, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Fatalf("expected %d records but got %d instead", 1, len(records))
	}
	r1 := records[0].Data.(tests.VehicleRegistered)
	if r1.License != e1.License ||
		r1.State != e1.State ||
		r1.VIN != e1.VIN {
		t.Fatal("embedded properties did not correctly deserialize")
	}

	tx := storage.GetTx[pgx.Tx](ctx)
	err = tx.Rollback(ctx)
	if err != nil {
		t.Error(err)
	}
}

func TestMapStorage(t *testing.T) {
	store := postgres.CreatePostgresStorage(
		CONNECTION_STRING,
	)

	ids1 := spry.Identifiers{"Name": "Gandalf", "Title": "The Grey"}
	ids2 := spry.Identifiers{"Name": "Gandalf", "Title": "The White"}
	aid, _ := storage.GetId()

	ctx, _ := store.GetContext(context.Background())
	err := store.AddMap(ctx, "Player", ids1, aid)
	if err != nil {
		t.Fatal("failed to add id map for id1", err)
	}
	err = store.AddMap(ctx, "Player", ids2, aid)
	if err != nil {
		t.Fatal("failed to add id map for id2", err)
	}

	read1, err := store.FetchId(ctx, "Player", ids1)
	if err != nil {
		t.Fatal("failed to read id for ids1", err)
	}
	read2, err := store.FetchId(ctx, "Player", ids2)
	if err != nil {
		t.Fatal("failed to read id for ids1", err)
	}

	if read1 != aid {
		t.Fatal("loaded the incorrect id for ids1")
	}
	if read2 != aid {
		t.Fatal("loaded the incorrect id for ids2")
	}

	tx := storage.GetTx[pgx.Tx](ctx)
	err = tx.Rollback(ctx)
	if err != nil {
		t.Error(err)
	}
}

func TestSnapshotStorage(t *testing.T) {
	store := postgres.CreatePostgresStorage(
		CONNECTION_STRING,
	)

	uid1, _ := storage.GetId()
	uid2, _ := storage.GetId()
	person1 := tests.Player{
		Name:      "Billy",
		HitPoints: 100,
		Dead:      false,
	}
	snap1 := storage.Snapshot{
		Id:            uid1,
		ActorId:       uid1,
		Type:          "Player",
		Version:       0,
		CreatedOn:     time.Now(),
		EventsApplied: 1,
		LastEventId:   uid1,
		LastCommandId: uid1,
		LastCommandOn: time.Now(),
		LastEventOn:   time.Now(),
		Data:          person1,
	}

	person2 := tests.Player{
		Name:      "Billy",
		HitPoints: 0,
		Dead:      true,
	}
	snap2 := storage.Snapshot{
		Id:            uid2,
		ActorId:       uid1,
		Type:          "Player",
		Version:       0,
		CreatedOn:     time.Now(),
		EventsApplied: 2,
		LastEventId:   uid2,
		LastCommandId: uid2,
		LastCommandOn: time.Now(),
		LastEventOn:   time.Now(),
		Data:          person2,
	}

	ctx, _ := store.GetContext(context.Background())
	err := store.AddSnapshot(ctx, "Player", snap1, true)
	if err != nil {
		t.Fatal("failed to persist snapshot 1", err)
	}
	err = store.AddSnapshot(ctx, "Player", snap2, true)
	if err != nil {
		t.Fatal("failed to persist snapshot 2", err)
	}

	latest, err := store.FetchLatestSnapshot(ctx, "player", uid1)
	if err != nil {
		t.Fatal("failed to read the latest snapshot for uuid")
	}
	if latest.ActorId != uid1 ||
		latest.Data == "" ||
		latest.EventsApplied != 2 ||
		latest.LastEventId != uid2 ||
		latest.LastCommandId != uid2 {
		t.Fatal("snapshot record did not load or deserialize correctly")
	}

	tx := storage.GetTx[pgx.Tx](ctx)
	err = tx.Rollback(ctx)
	if err != nil {
		t.Error(err)
	}
}
