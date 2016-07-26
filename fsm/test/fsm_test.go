package test

import (
	"fmt"
	"testing"

	"github.com/quintans/toolkit/fsm"
)

func TestSimpleTransition(t *testing.T) {
	sm := fsm.NewFSM()

	sm.For("GREEN").When("TICK", "YELLOW")
	sm.For("YELLOW").When("TICK", "RED")
	sm.For("RED").When("TICK", "GREEN")
	sm.Start()

	sm.Trigger("TICK")
	if sm.State() != "YELLOW" {
		t.Error("Expected state YELLOW got %s", sm.State())
	}

	sm.Trigger("TICK")
	if sm.State() != "RED" {
		t.Error("Expected state RED got %s", sm.State())
	}

	sm.Trigger("TICK")
	if sm.State() != "GREEN" {
		t.Error("Expected state GREEN got %s", sm.State())
	}
}

func TestTransitionWithEvents(t *testing.T) {

	onBeforeTransitionCnt := 0
	onExitCnt := 0
	onAllwaysCnt := 0
	//onEnterCnt := 0
	onAfterTransitionCnt := 0
	onStateChangeCnt := 0

	onBeforeTransition := func() { onBeforeTransitionCnt++ }
	onExit := func() { onExitCnt++ }
	onAllways := func() { onAllwaysCnt++; fmt.Println("onAllwaysCnt:", onAllwaysCnt) }
	//onEnter := func() { onEnterCnt++ }
	onAfterTransition := func() { onAfterTransitionCnt++ }
	onStateChange := func() { onStateChangeCnt++ }

	sm := fsm.NewFSM()
	sm.For("GREEN").
		When("TICK", "YELLOW"). // transition
		OnBeforeTransition(onBeforeTransition).
		OnAfterTransition(onAfterTransition).
		OnSet(onAllways).
		OnExit(onExit)
	sm.AdChangeListener(onStateChange)
	sm.Start()

	sm.Trigger("TICK")

	if sm.State() != "YELLOW" {
		t.Error("Expected state YELLOW got %s", sm.State())
	}

	if onBeforeTransitionCnt != 1 {
		t.Error("Expected 1 for onBeforeTransitionCnt, got %v", onBeforeTransitionCnt)
	}
}
