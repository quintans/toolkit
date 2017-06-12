package test

import (
	"testing"

	"github.com/quintans/toolkit/fsm2"
)

// event
const (
	BOING = "BOING"
	TICK  = "TICK"
	LOOP  = "LOOP"
)

// states
var (
	green  = fsm2.NewState("GREEN")
	yellow = fsm2.NewState("YELLOW")
	red    = fsm2.NewState("RED")
	bounce = fsm2.NewState("BOUNCE")
)

func TestSimpleTransition(t *testing.T) {
	// transitions
	green.AddTransition(TICK, yellow)
	yellow.AddTransition(TICK, bounce)
	bounce.AddTransition(BOING, red)
	bounce.OnEvent = func(e *fsm2.Event) *fsm2.Event {
		return fsm2.NewEvent(BOING, nil)
	}

	red.AddTransition(TICK, green)
	red.AddTransition(LOOP, red)
	var redState struct {
		ExitCount  int
		EnterCount int
		EventCount int
	}
	red.OnEnter = func(e *fsm2.Event) {
		redState.EnterCount++
	}
	red.OnExit = func(e *fsm2.Event) {
		redState.ExitCount++
	}
	red.OnEvent = func(e *fsm2.Event) *fsm2.Event {
		redState.EventCount++
		return nil
	}

	// Sate machine
	sm := fsm2.NewStateMachine("SimpleTransition")
	sm.AddState(green)
	sm.AddState(yellow)
	sm.AddState(red)

	sm.Event(TICK, nil)
	if sm.State() != yellow {
		t.Error("Expected state YELLOW got,", sm.State())
	}

	sm.Event(TICK, nil)
	if sm.State() != red {
		t.Error("Expected state RED got,", sm.State())
	}

	sm.Event(LOOP, nil)
	sm.Event(LOOP, nil)
	if sm.State() != red {
		t.Error("Expected state RED got,", sm.State())
	}
	if redState.EnterCount != 1 {
		t.Error("Expected RED OnEnter count of 1, got", redState.EnterCount)
	}
	if redState.EventCount != 3 {
		t.Error("Expected RED OnEvent count of 3, got", redState.EventCount)
	}
	if redState.ExitCount != 0 {
		t.Error("Expected RED OnExit count of 0, got", redState.ExitCount)
	}

	sm.Event(TICK, nil)
	if sm.State() != green {
		t.Error("Expected state GREEN got,", sm.State())
	}

	if redState.ExitCount != 1 {
		t.Error("Expected RED OnExit count of 1, got", redState.ExitCount)
	}

}
