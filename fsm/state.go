// Package fsm provides a flexible, thread-safe implementation of finite state machines.
// It allows for defining states, transitions between states, and conditions for those
// transitions. The implementation is context-aware and thread-safe, making it suitable
// for concurrent applications.
package fsm

import (
	"context"
	"sync/atomic"
)

//go:generate go run github.com/schigh/slice/cmd/slicify Transition all

// Global ID counter for generating unique IDs for states and transitions.
var gid uint64                    // nolint:gochecknoglobals

// Lock channel for thread-safe ID generation.
var lock = make(chan struct{}, 1) // nolint:gochecknoglobals

// resetGID resets the global ID counter to zero.
// This is primarily used for testing purposes.
func resetGID() {
	atomic.StoreUint64(&gid, 0)
}

// mkID generates a unique ID for states and transitions.
// It uses a channel-based lock to ensure thread safety.
func mkID() uint64 {
	lock <- struct{}{}
	defer func() {
		<-lock
	}()
	id := atomic.AddUint64(&gid, 1)
	return id
}

// Trigger is a type representing a trigger event in the state machine.
type Trigger uint

// TriggerFunc is a function type that evaluates whether a transition should occur.
// It takes a context for cancellation and a value to evaluate against.
// Returns:
// - bool: true if the transition should occur, false otherwise
// - error: any error that occurred during evaluation
//
// Example:
//
//	t := s1.When("value == 'a'", func(ctx context.Context, v interface{}) (bool, error) {
//	    if s, ok := v.(string); ok {
//	        return s == "a", nil
//	    }
//	    return false, nil
//	})
type TriggerFunc func(context.Context, interface{}) (bool, error)

// Identifier is an interface for objects with unique IDs.
// All states and transitions implement this interface.
type Identifier interface {
	// Id returns the unique identifier for this object.
	Id() uint64
}

// State represents a state in the finite state machine.
// Each state has a unique ID and name, and can create transitions to other states.
type State interface {
	Identifier
	// Name returns the name of the state.
	Name() string
	// When creates a new transition from this state with the specified condition.
	// The description parameter provides a human-readable description of the condition.
	// The TriggerFunc determines when the transition should occur.
	When(string, TriggerFunc) Transition
}

// Transition represents a transition between states in the finite state machine.
// Each transition has a source state, destination state, and a condition for when
// the transition should occur.
type Transition interface {
	Identifier
	// Description returns a human-readable description of the transition condition.
	Description() string
	// From returns the source state of the transition.
	From() State
	// To returns the destination state of the transition.
	To() State
	// Then sets the destination state of the transition and returns the transition.
	Then(State) Transition
	// Go evaluates whether the transition should occur based on the provided value.
	// Returns true if the transition should occur, false otherwise.
	Go(context.Context, interface{}) (bool, error)
}

// machineState is the concrete implementation of the State interface.
// It represents a state in the finite state machine with a name and unique ID.
type machineState struct {
	name string  // The name of the state
	id   uint64  // The unique identifier for the state
}

// StateOption is a function type used to configure a machineState.
// It follows the functional options pattern for configuring states.
type StateOption func(machineState) machineState

// NewState creates a new state with the specified name and options.
// Each state has a unique ID generated automatically.
//
// Example:
//
//	s1 := fsm.NewState("STATE1")
//	s2 := fsm.NewState("STATE2")
func NewState(name string, options ...StateOption) State {
	s := machineState{id: mkID(), name: name}
	for _, f := range options {
		s = f(s)
	}

	return s
}

// Name returns the name of the state.
func (s machineState) Name() string {
	return s.name
}

// When creates a new transition from this state with the specified condition.
// The description parameter provides a human-readable description of the condition.
// The TriggerFunc determines when the transition should occur.
//
// Example:
//
//	t := s1.When("value == 'a'", func(ctx context.Context, v interface{}) (bool, error) {
//	    if s, ok := v.(string); ok {
//	        return s == "a", nil
//	    }
//	    return false, nil
//	}).Then(s2)
func (s machineState) When(desc string, f TriggerFunc) Transition {
	return &edge{id: mkID(), from: s, f: f, desc: desc}
}

// Id returns the unique identifier for this state.
func (s machineState) Id() uint64 {
	return s.id
}

// edge is the concrete implementation of the Transition interface.
// It represents a transition between states in the finite state machine.
type edge struct {
	desc string      // Human-readable description of the transition condition
	from State       // Source state of the transition
	to   State       // Destination state of the transition
	f    TriggerFunc // Function that determines when the transition should occur
	id   uint64      // Unique identifier for the transition
}

// Id returns the unique identifier for this transition.
func (e *edge) Id() uint64 {
	return e.id
}

// Description returns a human-readable description of the transition condition.
func (e *edge) Description() string {
	return e.desc
}

// From returns the source state of the transition.
func (e *edge) From() State {
	return e.from
}

// To returns the destination state of the transition.
func (e *edge) To() State {
	return e.to
}

// Then sets the destination state of the transition and returns the transition.
// This method is part of the fluent API for creating transitions.
//
// Example:
//
//	t := s1.When("condition", conditionFunc).Then(s2)
func (e *edge) Then(s State) Transition {
	e.to = s
	return e
}

// Go evaluates whether the transition should occur based on the provided value.
// It delegates to the TriggerFunc associated with this transition.
// Returns true if the transition should occur, false otherwise.
func (e *edge) Go(ctx context.Context, v interface{}) (bool, error) {
	return e.f(ctx, v)
}
