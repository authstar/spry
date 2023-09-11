package tests

import (
	"testing"

	"github.com/authstar/spry"
	"github.com/authstar/spry/core"
)

func TestIdSetBehavior(t *testing.T) {
	ids := spry.CreateIdSet()
	set := ids.ToIdentifierSet()

	if len(set) > 0 {
		t.Errorf("length of id set should be 0")
	}

	id1 := spry.Identifiers{"Name": "Tina"}
	s1, _ := spry.IdentifiersToString(id1)

	ids.AddIdsFor("Player", id1)
	set = ids.ToIdentifierSet()

	if len(set) != 1 {
		t.Errorf("the length of id set should be 0")
	}

	if len(set["Player"]) != 1 {
		t.Error("number of ids in id set is incorrect")
	}

	s, _ := spry.IdentifiersToString(ids.GetIdsFor("Player")[0])
	if s != s1 {
		t.Error("ids in map do not match expectation")
	}

	id2 := spry.Identifiers{"Name": "Amy"}
	id3 := spry.Identifiers{"Name": "Casey"}
	id4 := spry.Identifiers{"Name": "Mia"}
	id5 := spry.Identifiers{"Name": "Melissa"}
	id6 := spry.Identifiers{"Name": "Jason"}
	id7 := spry.Identifiers{"Name": "Andy"}
	id8 := spry.Identifiers{"Name": "Bill"}
	id9 := spry.Identifiers{"Name": "Will"}
	id10 := spry.Identifiers{"Name": "Chris"}

	s2, _ := spry.IdentifiersToString(id2)
	s3, _ := spry.IdentifiersToString(id3)
	s4, _ := spry.IdentifiersToString(id4)
	s5, _ := spry.IdentifiersToString(id5)
	s6, _ := spry.IdentifiersToString(id6)
	s7, _ := spry.IdentifiersToString(id7)
	s8, _ := spry.IdentifiersToString(id8)
	s9, _ := spry.IdentifiersToString(id9)
	s10, _ := spry.IdentifiersToString(id10)

	ids.AddIdsFor("Player", id2, id3, id4, id5)
	ids.AddIdsFor("Player", id6, id7, id8, id9, id10)

	list := ids.GetIdsFor("Player")
	if len(list) != 10 {
		t.Error("ids missing from list")
	}

	idstrings := IdsToStrings(list)
	correct := []string{s1, s2, s3, s4, s5, s6, s7, s8, s9, s10}
	matching := AreEqual(correct, idstrings)
	if !matching {
		t.Error("ids for type did not match expectations")
	}

	ids.RemoveIdsFrom("Player", spry.Identifiers{"Name": "Will"})
	list = ids.GetIdsFor("Player")
	idstrings = IdsToStrings(list)
	correct = []string{s1, s2, s3, s4, s5, s6, s7, s8, s10}
	matching = AreEqual(correct, idstrings)
	if !matching {
		t.Error("ids for type did not match expectations")
	}

	ids.RemoveIdsFrom("Player", spry.Identifiers{"Name": "Mia"})
	list = ids.GetIdsFor("Player")
	idstrings = IdsToStrings(list)
	correct = []string{s1, s2, s3, s5, s6, s7, s8, s10}
	matching = AreEqual(correct, idstrings)
	if !matching {
		t.Error("ids for type did not match expectations")
	}

	ids.RemoveIdsFrom("Player", spry.Identifiers{"Name": "Chris"})
	list = ids.GetIdsFor("Player")
	idstrings = IdsToStrings(list)
	correct = []string{s1, s2, s3, s5, s6, s7, s8}
	matching = AreEqual(correct, idstrings)
	if !matching {
		t.Error("ids for type did not match expectations")
	}
}

func AreEqual(correct []string, idstrings []string) bool {
	matching := core.Reducer(correct, func(m bool, id string, i int) bool {
		return m && id == idstrings[i]
	}, true)
	return matching
}

func IdsToStrings(list []spry.Identifiers) []string {
	return core.Mapper(
		list,
		func(x spry.Identifiers) string {
			s, _ := spry.IdentifiersToString(x)
			return s
		},
	)
}
