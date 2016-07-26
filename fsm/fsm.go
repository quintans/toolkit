package fsm

// copied and adapted from
// http://www.java2s.com/Code/Java/Collections-Data-Structure/AprogrammableFiniteStateMachineimplementation.htm

// $Id: FSM.java 12 2009-11-09 22:58:47Z gabe.johnson $

//  A programmable Finite State Machine implementation. To use this class,
//  establish any number of states with the 'addState' method. Next, add some
//  FSM.Transition objects (the Transition class is designed to be used as an
//  superclass for your anonymous implementation). Each Transition object has two
//  useful methods that can be defined by your implementation: doBeforeTransition
//  and doAfterTransition. To drive your FSM, simply give it events using the
//  addEvent method with the name of an event. If there is an appropriate
//  transition for the current state, the transition's doBefore/doAfter methods
//  are called and the FSM is put into the new state. It is legal (and highly
//  useful) for the start/end states of a transition to be the same state.

type FSM struct {
	initialState    *State
	currentState    *State
	states          map[string]*State
	changeListeners []func()
}

func NewFSM() *FSM {
	this := new(FSM)
	this.states = make(map[string]*State)
	this.changeListeners = make([]func(), 0)
	return this
}

// Report the current state of the finite state machine.
func (this *FSM) State() string {
	if this.currentState != nil {
		return this.currentState.name
	} else {
		return ""
	}
}

// Establish a new state the FSM is aware of. If the FSM does not currently
// have any states, this state becomes the current, initial state. This is
// the only way to put the FSM into an initial state.
//
// The entryCode, exitCode, and alwaysRunCode are Runnables that the FSM
// executes during the course of a transition. entryCode and exitCode are
// run only if the transition is between two distinct states (i.e. A->B
// where A != B). alwaysRunCode is executed even if the transition is
// re-entrant (i.e. A->B where A = B).
func (this *FSM) For(state string) *State {
	s, _ := this.states[state]
	if s == nil {
		isInitial := len(this.states) == 0
		s = newState(this, state)
		this.states[state] = s
		if isInitial {
			this.initialState = s
		}
	}
	return s
}

func (this *FSM) Start() {
	if this.initialState != nil {
		this.setStateChange(this.initialState, true)
	}
}

// SetState sets the current state without following a transition.
// This will cause a change event to be fired.
func (this *FSM) SetState(state string) {
	this.setStateChange(this.For(state), true)
}

// Sets the current state without followign a transition, and optionally
// causing a change event to be triggered. During state transitions (with
// the 'addEvent' method), this method is used with the triggerEvent
// parameter as false.
//
// The FSM executes non-null runnables according to the following logic,
// given start and end states A and B:
//
// - If A and B are distinct, run A's exit code.
// - Record current state as B.
// - Run B's "alwaysRunCode".
// - If A and B are distinct, run B's entry code.
func (this *FSM) SetStateChange(state string, triggerChange bool) {
	this.setStateChange(this.For(state), triggerChange)
}

func (this *FSM) setStateChange(state *State, triggerChange bool) {
	runExtraCode := state != this.currentState
	if runExtraCode && this.currentState != nil {
		this.currentState.runOnExit()
	}
	this.currentState = state
	this.currentState.runOnSet()
	if runExtraCode {
		this.currentState.runOnEnter()
	}
	if triggerChange {
		this.fireChangeEvent()
	}
}

// Add a change listener.
// Is only used to report changes that have already happened. ChangeEvents are
// only fired AFTER a transition's doAfterTransition is called.
func (this *FSM) AdChangeListener(listener func()) {
	this.changeListeners = append(this.changeListeners, listener)
}

// Fire a change event to registered listeners.
func (this *FSM) fireChangeEvent() {
	for _, v := range this.changeListeners {
		v()
	}
}

// Feed the FSM with the named event. If the current state has a transition
// that responds to the given event, the FSM will performed the transition
// using the following steps, assume start and end states are A and B:
//
// - Execute the transition's "OnBeforeTransition" method
// - Run fsm.SetStateChange(B, false) -- see docs for that method
// - Execute the transition's "OnAfterTransition" method
// - Fire a change event, notifying interested observers that the transition has completed.
// - Now firmly in state B, see if B has a third state C that we must automatically transition to via TriggerEvent(C).
func (this *FSM) Trigger(evtName string) string {
	state := this.currentState

	if trans, ok := state.transitions[evtName]; ok {
		if trans.onBeforeTransition != nil {
			trans.onBeforeTransition()
		}
		this.setStateChange(trans.endState, false)
		if trans.onAfterTransition != nil {
			trans.onAfterTransition()
		}
		this.fireChangeEvent()
		forward := trans.endState.forwardState
		if forward != nil {
			this.Trigger(forward())
		}
	}

	return this.currentState.name
}

//  state represents a state with some number of associated transitions.
type State struct {
	sm             *FSM // state machine
	name           string
	lastTransition *Transition
	transitions    map[string]*Transition
	// There are cases where a state is meant to be transitional, and the FSM
	// should always immediately transition to some other state. In those cases,
	// use this method to specify the end states. After the startState
	// has fully transitioned (and any change events have been fired) the FSM
	// will check to see if there is another state that the FSM should
	// automatically transition to. If there is one, addEvent(endState) is
	// called.
	forwardState func() string
	onEnter      func()
	onExit       func()
	onSet        func()
}

func newState(sm *FSM, name string) *State {
	this := new(State)
	this.sm = sm
	this.name = name
	this.transitions = make(map[string]*Transition)
	return this
}

func (this *State) ForwardState(on func() string) *State {
	this.forwardState = on
	return this
}

func (this *State) OnEnter(on func()) *State {
	this.onEnter = on
	return this
}

func (this *State) OnExit(on func()) *State {
	this.onExit = on
	return this
}

func (this *State) OnSet(on func()) *State {
	this.onSet = on
	return this
}

func (this *State) OnBeforeTransition(on func()) *State {
	this.lastTransition.onBeforeTransition = on
	return this
}

func (this *State) OnAfterTransition(on func()) *State {
	this.lastTransition.onAfterTransition = on
	return this
}

// When creates the transition with the name 'event'
func (this *State) When(event string, endState string) *State {
	t := this.transitions[event]
	if t == nil {
		t = newTransition(event, this, this.sm.For(endState))
		this.transitions[event] = t
	}
	this.lastTransition = t
	return this
}

func (this *State) runOnEnter() {
	if this.onEnter != nil {
		this.onEnter()
	}
}

func (this *State) runOnExit() {
	if this.onExit != nil {
		this.onExit()
	}
}

func (this *State) runOnSet() {
	if this.onSet != nil {
		this.onSet()
	}
}

type Transition struct {
	evtName    string
	startState *State
	endState   *State
	// Set this to have FSM execute code immediately before following a state transition.
	onBeforeTransition func()
	// Set this to have FSM execute code immediately after following a state transition.
	onAfterTransition func()
}

// NewTransition creates a transition object that responds to the given event when in
// the given startState, and puts the FSM into the endState provided.
func newTransition(evtName string, startState *State, endState *State) *Transition {
	this := new(Transition)
	this.evtName = evtName
	this.startState = startState
	this.endState = endState
	return this
}

/*
func (this *Transition) OnBeforeTransition(on func()) *Transition {
	this.onBeforeTransition = on
	return this
}

func (this *Transition) OnAfterTransition(on func()) *Transition {
	this.onAfterTransition = on
	return this
}
*/
