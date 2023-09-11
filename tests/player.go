package tests

import "github.com/authstar/spry"

// an aggregate
type World struct {
	Name        string
	PlayerCount int
	Players     []string
}

func (w World) GetIdentifiers() spry.Identifiers {
	return map[string]any{"name": w.Name}
}

// Player actor
type Player struct {
	Name      string
	HitPoints int
	Dead      bool
}

func (p Player) GetIdentifiers() spry.Identifiers {
	return map[string]any{"name": p.Name}
}

// events

type PlayerCreated struct {
	Name string
}

func (event PlayerCreated) applyToPlayer(player *Player) {
	player.HitPoints = 100
	player.Name = event.Name
}

func (event PlayerCreated) applyToWorld(world *World) {
	world.Players = append(world.Players, event.Name)
	world.PlayerCount++
}

func (event PlayerCreated) Apply(actor any) any {
	switch a := actor.(type) {
	case *Player:
		event.applyToPlayer(a)
	case *World:
		event.applyToWorld(a)
	}
	return actor
}

type PlayerDamaged struct {
	Damage int
}

func (event PlayerDamaged) Apply(actor any) any {
	switch a := actor.(type) {
	case *Player:
		a.HitPoints -= event.Damage
	}
	return actor
}

type PlayerHealed struct {
	Health int
}

func (event PlayerHealed) Apply(actor any) any {
	switch a := actor.(type) {
	case *Player:
		a.HitPoints += event.Health
	}
	return actor
}

type PlayerDied struct {
	Message string
}

func (event PlayerDied) Apply(actor any) any {
	switch a := actor.(type) {
	case *World:
		a.PlayerCount--
	}
	return actor
}

// commands

type CreatePlayer struct {
	Name string
}

func (command CreatePlayer) GetIdentifiers() spry.Identifiers {
	return spry.Identifiers{"name": command.Name}
}

func (command CreatePlayer) Handle(actor any) ([]spry.Event, []error) {
	var events []spry.Event
	switch actor.(type) {
	case Player:
		events = append(events, PlayerCreated(command))
	}
	return events, []error{}
}

type DamagePlayer struct {
	Name   string
	Damage int
}

func (command DamagePlayer) GetIdentifiers() spry.Identifiers {
	return spry.Identifiers{"name": command.Name}
}

func (command DamagePlayer) Handle(actor any) ([]spry.Event, []error) {
	var events []spry.Event
	switch a := actor.(type) {
	case Player:
		if a.HitPoints <= command.Damage {
			events = append(events, PlayerDied{Message: "you died"})
		}
		events = append(events, PlayerDamaged{Damage: command.Damage})
	}
	return events, []error{}
}

type HealPlayer struct {
	Name   string
	Health int
}

func (command HealPlayer) GetIdentifiers() spry.Identifiers {
	return spry.Identifiers{"name": command.Name}
}

func (command HealPlayer) Handle(actor any) ([]spry.Event, []error) {
	var events []spry.Event
	switch actor.(type) {
	case Player:
		events = append(events, PlayerHealed{Health: command.Health})
	}
	return events, []error{}
}
