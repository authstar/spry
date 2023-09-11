package tests

import (
	"testing"
	"time"

	"github.com/authstar/spry"
	"github.com/authstar/spry/storage"
	"github.com/mitchellh/mapstructure"
)

func TestEventRecordSerialization(t *testing.T) {
	uid, err := storage.GetId()
	if err != nil {
		t.Fatal("could not create UUID")
	}
	buffer, err := spry.ToJson(storage.EventRecord{
		Id:               uid,
		Type:             "ThingHappened",
		ActorId:          uid,
		ActorName:        "Test",
		CreatedOn:        time.Now(),
		CreatedByVersion: 0,
		CreatedBy:        "Test",
		CreatedById:      uid,
	})
	if err != nil {
		t.Fatal("failed to serialize record", err)
	}

	record, err := spry.FromJson[storage.EventRecord](buffer)
	if err != nil {
		t.Fatal("failed to deserialize record from json", err)
	}
	if record.Id != uid ||
		record.Type != "ThingHappened" ||
		record.ActorId != uid ||
		record.ActorName != "Test" ||
		record.CreatedByVersion != 0 ||
		record.CreatedBy != "Test" ||
		record.CreatedById != uid {
		t.Fatal("deserialization created invalid object", err)
	}
}

func TestEmbeddedSerialization(t *testing.T) {
	vr1 := VehicleRegistered{
		MotoristId: MotoristId{License: "123", State: "AK"},
		VehicleId:  VehicleId{VIN: "000000001"},
		Type:       "scooter",
		Make:       "gogeddums",
		Model:      "scootchies",
		Color:      "arglebargle",
	}
	json, err := spry.ToJson(vr1)
	if err != nil {
		t.Error(err)
	}

	tmp, err := spry.FromJson[any](json)

	var vr2 VehicleRegistered
	_ = mapstructure.Decode(tmp, &vr2)

	if err != nil {
		t.Error(err)
	}

	if vr2.License != vr1.License ||
		vr2.State != vr1.State ||
		vr2.VIN != vr1.VIN ||
		vr2.Color != vr1.Color ||
		vr2.Make != vr1.Make ||
		vr2.Model != vr1.Model ||
		vr2.Type != vr1.Type {
		t.Error("embedded fields are not correctly preserved during JSON serialization")
	}
}
