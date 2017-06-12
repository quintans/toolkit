package fsm

type StateMachine struct {
	name            string
	states          map[string]*State
	currentState    *State
	changeListeners []func(*Event)
}

func NewStateMachine(name string) *StateMachine {
	this := new(StateMachine)
	this.name = name
	this.states = make(map[string]*State)
	this.changeListeners = make([]func(*Event), 0)
	return this
}

// AddState adds state to the StateMachine.
// If it is the first state to be add, it will be the initial state
func (this *StateMachine) AddState(state *State) {
	this.states[state.name] = state
	if this.currentState == nil {
		this.SetState(state, nil)
	}
}

func (this *StateMachine) SetState(state *State, event *Event) *Event {
	var diffState = state != this.currentState
	if diffState && this.currentState != nil && this.currentState.OnExit != nil {
		this.currentState.OnExit(event)
	}
	this.currentState = state
	var nextEvent *Event
	if state.OnEvent != nil {
		nextEvent = this.currentState.OnEvent(event)
	}
	if diffState && this.currentState.OnEnter != nil {
		this.currentState.OnEnter(event)
	}

	if event != nil {
		this.fireChangeEvent(event)
	}

	return nextEvent
}

func (this *StateMachine) Event(name string, data interface{}) {
	var state = this.currentState
	if endState, ok := state.transitions[name]; ok {
		var event = &Event{name, data, state}
		var nextEvent = this.SetState(endState, event)
		this.fireChangeEvent(event)
		if nextEvent != nil {
			this.Event(nextEvent.name, nextEvent.data)
		}
	}
}

func (this *StateMachine) State() *State {
	return this.currentState
}

func (this *StateMachine) String() string {
	return this.name
}

// Add a change listener.
// Is only used to report changes that have already happened. ChangeEvents are
// only fired AFTER a transition's doAfterTransition is called.
func (this *StateMachine) AdChangeListener(listener func(*Event)) {
	this.changeListeners = append(this.changeListeners, listener)
}

// Fire a change event to registered listeners.
func (this *StateMachine) fireChangeEvent(event *Event) {
	for _, v := range this.changeListeners {
		v(event)
	}
}

type State struct {
	name        string
	transitions map[string]*State
	// OnEnter is called when entering a state
	// when there is a transition A -> B where A != B
	OnEnter func(*Event)
	// OnExit is called when exiting a state
	// when there is a transition A -> B where A != B
	OnExit func(*Event)
	// OnEvent is called when a event occurrs, even if
	// the transition A -> B where A == B.
	// An event can be returned in the case of a transitional state.
	OnEvent func(*Event) *Event
}

func NewState(name string) *State {
	this := new(State)
	this.name = name
	this.transitions = make(map[string]*State)
	return this
}

// AddTransition adds a state transition.
func (this *State) AddTransition(event string, to *State) *State {
	this.transitions[event] = to
	return this
}

func (this *State) Name() string {
	return this.name
}

func (this *State) String() string {
	return this.name
}

type Event struct {
	name string
	data interface{}
	from *State
}

func (this *Event) Name() string {
	return this.name
}

func (this *Event) Data() interface{} {
	return this.data
}

func (this *Event) From() *State {
	return this.from
}

func NewEvent(name string, data interface{}) *Event {
	return &Event{name, data, nil}
}
