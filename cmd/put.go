// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
)

// Syntax: PUT item container
func init() {
	AddHandler(Put, "PUT")
}

func Put(s *state) {

	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("You go to put something into something else...")
		return
	}

	tName := s.words[0]

	// Search ourselves for item to put into container
	tWhere := attr.FindInventory(s.actor)
	tWhat := tWhere.Search(tName)

	if tWhat == nil {
		s.msg.Actor.SendBad("You have no '", tName, "' to put into anything.")
		return
	}

	// Get item's proper name
	tName = attr.FindName(tWhat).Name(tName)

	// Check a container was specified
	if len(s.words) < 2 {
		s.msg.Actor.SendBad("What did you want to put ", tName, " into?")
		return
	}

	cName := s.words[1]

	// Search ourselves for container to put something into
	cWhat := tWhere.Search(cName)

	// If container not found and we are not somewhere we're not going to find it
	if cWhat == nil && s.where == nil {
		s.msg.Actor.SendBad("There is no '", cName, "' to put ", tName, " into.")
		return
	}

	// If container not found search the inventory where we are
	if cWhat == nil {
		cWhat = s.where.Search(cName)
	}

	// Was container found?
	if cWhat == nil {
		s.msg.Actor.SendBad("You see no '", cName, "' to put ", tName, " into.")
		return
	}

	who := attr.FindName(s.actor).Name("Someone")

	// Unless our name is Klein we can't put something inside itself! ;)
	if tWhat == cWhat {
		s.msg.Actor.SendInfo("It might be interesting putting ", tName, " inside itself, but probably paradoxical as well.")
		s.msg.Observer.SendInfo(who, " seems to be trying to turn ", tName, " into a paradox.")
		return
	}

	// Get container's proper name
	cName = attr.FindName(cWhat).Name(cName)

	// Check container is actually a container with an inventory
	cWhere := attr.FindInventory(cWhat)
	if !cWhere.Found() {
		s.msg.Actor.SendBad("You cannot put ", tName, " into ", cName, ".")
		return
	}

	// Check for veto on item being put into container
	if veto := attr.FindVetoes(tWhat).Check("DROP", "PUT"); veto != nil {
		s.msg.Actor.SendBad(veto.Message())
		return
	}

	// Check for veto on container
	if veto := attr.FindVetoes(cWhat).Check("PUT"); veto != nil {
		s.msg.Actor.SendBad(veto.Message())
		return
	}

	// Remove item from where it is and put it in the container
	if tWhere.Move(tWhat, cWhere) == nil {
		s.msg.Actor.SendBad("Something stops you putting ", tName, " anywhere.")
		return
	}

	s.msg.Actor.SendGood("You put ", tName, " into ", cName, ".")

	s.msg.Observer.SendInfo("You see ", who, " put something into ", cName, ".")
	s.ok = true
}