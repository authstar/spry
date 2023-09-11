package storage

import (
	"fmt"
	"reflect"

	"github.com/authstar/spry"
	"github.com/mitchellh/mapstructure"
)

type Caster = func(any) (any, error)

type TypeMap struct {
	Events   map[string]Caster
	Commands map[string]Caster
}

func (m TypeMap) getCaster(t any) Caster {
	return func(v any) (any, error) {
		// target := reflect.New(t)
		err := mapstructure.Decode(v, &t)
		if err != nil {
			return t, err
		}
		return t, nil
	}
}

func (m TypeMap) AddTypes(types ...any) {
	for _, i := range types {
		it := reflect.TypeOf(i)
		name := it.Name()
		switch i.(type) {
		case spry.Event:
			m.Events[name] = m.getCaster(i)
		case spry.Command:
			m.Commands[name] = m.getCaster(i)
		}
	}
}

func (m TypeMap) AsEvent(eventType string, v any) (spry.Event, error) {
	if converter, ok := m.Events[eventType]; ok {
		e, err := converter(v)
		if err != nil {
			return nil, err
		}
		return e.(spry.Event), nil
	}
	return nil, fmt.Errorf("%s is an unregistered event", eventType)
}

func (m TypeMap) AsCommand(commandType string, v any) (spry.Command, error) {
	converter := m.Commands[commandType]
	c, err := converter(v)
	if err != nil {
		return nil, err
	}
	return c.(spry.Command), nil
}

func CreateTypeMap() TypeMap {
	return TypeMap{
		Events:   map[string]Caster{},
		Commands: map[string]Caster{},
	}
}
