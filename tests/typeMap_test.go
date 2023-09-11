package tests

import (
	"testing"

	"github.com/authstar/spry/storage"
)

func TestEventMapping(t *testing.T) {
	m := storage.CreateTypeMap()
	m.AddTypes(PlayerCreated{}, PlayerDamaged{}, PlayerHealed{})

	var i1 any = map[string]any{"Name": "Bob"}
	e1, _ := m.AsEvent("PlayerCreated", i1)
	var a1 Player = *new(Player)

	e1.Apply(&a1)
	if a1.Name != "Bob" {
		t.Error("failed to correctly construct event")
	}
}

func TestCommandMapping(t *testing.T) {
	m := storage.CreateTypeMap()
	m.AddTypes(CreatePlayer{}, DamagePlayer{}, HealPlayer{})

	var i1 any = map[string]any{"Name": "Bob", "Damage": 40}
	c1, _ := m.AsCommand("DamagePlayer", i1)
	var a1 Player = Player{Name: "Bob", HitPoints: 100}

	events, _ := c1.Handle(a1)
	if events[0].(PlayerDamaged).Damage != 40 {
		t.Error("failed to correctly construct event")
	}
}
