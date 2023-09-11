package tests

import (
	"testing"

	"github.com/authstar/spry"
)

type Withit struct {
	Name string
}

func (w Withit) GetIdentifiers() spry.Identifiers {
	return spry.Identifiers{"Name": w.Name}
}

func (w Withit) GetActorMeta() spry.ActorMeta {
	return spry.ActorMeta{
		SnapshotFrequency:       100,
		SnapshotDuringRead:      true,
		SnapshotDuringWrite:     false,
		SnapshotDuringPartition: false,
	}
}

type Without struct {
	Name string
}

func (w Without) GetIdentifiers() spry.Identifiers {
	return spry.Identifiers{"Name": w.Name}
}

func TestActorWithConfig(t *testing.T) {
	meta := spry.GetActorMeta[Withit]()
	if meta.SnapshotFrequency != 100 ||
		meta.SnapshotDuringPartition == true ||
		meta.SnapshotDuringRead == false ||
		meta.SnapshotDuringWrite == true {
		t.Error("actor meta for withit did not match expected settings")
	}
}

func TestActorWithoutConfig(t *testing.T) {
	meta := spry.GetActorMeta[Without]()
	if meta.SnapshotFrequency != 20 ||
		meta.SnapshotDuringPartition == false ||
		meta.SnapshotDuringRead == true ||
		meta.SnapshotDuringWrite == false {
		t.Error("actor meta for withit did not match expected settings")
	}
}
