// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
)

// Syntax: PUT item container
func init() {
	AddHandler(Put, "PUT")
}

func Put(s *state) {

	if len(s.words) == 0 {
		s.msg.actor.WriteString("You go to put something into something else...")
		return
	}

	tName := s.words[0]

	// Search ourselves for item to put into container
	tWhere := attr.FindInventory(s.actor)
	tWhat := tWhere.Search(tName)

	if tWhat == nil {
		s.msg.actor.WriteJoin("You have no '", tName, "' to put into anything.")
		return
	}

	// Get item's proper name
	if n := attr.FindName(tWhat); n != nil {
		tName = n.Name()
	}

	// Check a container was specified
	if len(s.words) < 2 {
		s.msg.actor.WriteJoin("What did you want to put ", tName, " into?")
		return
	}

	// Try and find container
	var (
		cName = s.words[1]

		cWhat has.Thing
	)

	// Search ourselves for container to put something into
	cWhat = tWhere.Search(cName)

	// If container not found and we are not somewhere we are not going to find
	// it
	if cWhat == nil && s.where == nil {
		s.msg.actor.WriteJoin("There is no '", cName, "' to put ", tName, " into.")
		return
	}

	// If container not found search the inventory where we are
	if cWhat == nil {
		cWhat = s.where.Search(cName)
	}

	// If container still not found check narratives where we are
	if cWhat == nil {
		if a := attr.FindNarrative(s.where.Parent()); a != nil {
			cWhat = a.Search(cName)
		}
	}

	// Was container found?
	if cWhat == nil {
		s.msg.actor.WriteJoin("You see no '", cName, "' to put ", tName, " into.")
		return
	}

	// Unless our name is Klein we can't put something inside itself! ;)
	if tWhat == cWhat {
		s.msg.actor.WriteJoin("You can't put ", tName, " inside itself!")
		return
	}

	// Get container's proper name
	if n := attr.FindName(cWhat); n != nil {
		cName = n.Name()
	}

	// Check container is actually a container with an inventory
	cInv := attr.FindInventory(cWhat)
	if cInv == (*attr.Inventory)(nil) {
		s.msg.actor.WriteJoin("You cannot put ", tName, " into ", cName, ".")
		return
	}

	// Check for veto on item being put into container
	if vetoes := attr.FindVetoes(tWhat); vetoes != nil {
		if veto := vetoes.Check("DROP", "PUT"); veto != nil {
			s.msg.actor.WriteString(veto.Message())
			return
		}
	}

	// Check for veto on container
	if vetoes := attr.FindVetoes(cWhat); vetoes != nil {
		if veto := vetoes.Check("PUT"); veto != nil {
			s.msg.actor.WriteString(veto.Message())
			return
		}
	}

	// Remove item from where it is
	if tWhere.Remove(tWhat) == nil {
		s.msg.actor.WriteJoin("Something stops you putting ", tName, " anywhere.")
		return
	}

	// Put item into comtainer
	cInv.Add(tWhat)

	s.msg.actor.WriteJoin("You put ", tName, " into ", cName, ".")
	s.ok = true
}
