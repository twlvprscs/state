// Package fsm provides a flexible, thread-safe implementation of finite state machines.
// It allows for defining states, transitions between states, and conditions for those
// transitions. The implementation is context-aware and thread-safe, making it suitable
// for concurrent applications.
//
// Basic usage:
//
//	// Create states
//	s1 := fsm.NewState("STATE1")
//	s2 := fsm.NewState("STATE2")
//
//	// Define transitions
//	t1 := s1.When("condition", func(ctx context.Context, v interface{}) (bool, error) {
//		// Evaluate condition
//		return true, nil
//	}).Then(s2)
//
//	// Create Machine with transitions
//	Machine := fsm.NewMachine(fsm.WithTransitions(t1))
//
//	// Use the Machine
//	changed, err := Machine.Update(ctx, someValue)
package fsm

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/schigh/slice"
)

// Machine represents a finite state Machine with states and transitions.
// It maintains the current state, start state, end states, and transitions
// between states. All operations on the Machine are thread-safe.
type Machine struct {
	mu          sync.RWMutex            // Mutex for thread-safety
	curr        atomic.Value            // Current state
	start       atomic.Value            // Start state
	endStates   map[uint64]State        // Map of end state IDs to states
	idx         uint32                  // Index counter
	transitions map[uint64][]Transition // Map of state IDs to transitions
	cancel      func()                  // Cancellation function
}

// Option is a function type used to configure a Machine.
// It follows the functional options pattern for configuring the FSM.
type Option func(*Machine)

// WithTransitions creates an Option that adds the specified transitions to the Machine.
// If transitions are provided, the first state in the first transition is set as the
// start state of the Machine.
//
// Example:
//
//	s1 := fsm.NewState("STATE1")
//	s2 := fsm.NewState("STATE2")
//	t := s1.When("condition", conditionFunc).Then(s2)
//	Machine := fsm.NewMachine(fsm.WithTransitions(t))
func WithTransitions(transitions ...Transition) Option {
	return func(m *Machine) {
		m.mu.Lock()
		defer m.mu.Unlock()

		if m.transitions == nil {
			m.transitions = make(map[uint64][]Transition)
		}

		// this sets the first state in the first transition as the root
		if len(transitions) > 0 {
			m.start.Store(transitions[0].From())
		}

		for _, t := range transitions {
			id := t.From().Id()
			m.transitions[id] = append(m.transitions[id], t)
		}
	}
}

// NewMachine creates a new finite state Machine with the specified options.
// Options can be used to configure the Machine, such as adding transitions.
//
// Example:
//
//	Machine := fsm.NewMachine(fsm.WithTransitions(t1, t2))
func NewMachine(opts ...Option) *Machine {
	m := Machine{}
	for _, f := range opts {
		f(&m)
	}

	return &m
}

// Graph generates a visual representation of the state Machine.
// This is a placeholder for future implementation.
//func (m *Machine) Graph() {
//	m.mu.RLock()
//	defer m.mu.RUnlock()
//
//	// TODO: traverse the fsm and generate graph
//}

// SetStart sets the start state of the Machine by name.
// It returns an error if no state with the given name is found.
// The start state is also set as the current state.
//
// Example:
//
//	if err := Machine.SetStart("IDLE"); err != nil {
//	    // handle error
//	}
func (m *Machine) SetStart(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var start State
	for _, t := range m.transitions {
		tt := TransitionSlice(t)
		tName := tt.Filter(func(tr Transition) bool {
			return tr.From().Name() == name
		})
		if len(tName) > 0 {
			start = tName[0].From()
			break
		}
	}

	if start == nil {
		return fmt.Errorf("no state found with name: %s", name)
	}

	m.start.Store(start)
	m.curr.Store(start)

	return nil
}

// Reset resets the Machine to its start state.
// It returns an error if the Machine has no start state.
//
// Example:
//
//	if err := Machine.Reset(); err != nil {
//	    // handle error
//	}
func (m *Machine) Reset() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	starti := m.start.Load()
	if starti == nil {
		return errors.New("this machine has no start state")
	}
	start, _ := starti.(State)
	m.curr.Store(start)

	return nil
}

// SetEndStates sets the end states of the Machine by name.
// It returns an error if any of the specified state names are not found.
// End states are used to determine when the Machine has reached a terminal state.
//
// Example:
//
//	if err := Machine.SetEndStates("SUCCESS", "FAILURE"); err != nil {
//	    // handle error
//	}
func (m *Machine) SetEndStates(names ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.endStates == nil {
		m.endStates = make(map[uint64]State)
	}

	validNames := make(map[string]bool)

	for _, t := range m.transitions {
		tt := TransitionSlice(t)
		tt.Filter(func(tr Transition) bool {
			return slice.String(names).Contains(tr.To().Name())
		}).Each(func(tr Transition) {
			m.endStates[tr.To().Id()] = tr.To()
			validNames[tr.To().Name()] = true
		})
	}

	idx, ok := slice.String(names).IfEach(func(s string) bool {
		_, ok := validNames[s]
		return ok
	})

	if !ok {
		return fmt.Errorf("invalid state: '%s'", names[idx])
	}

	return nil
}

// IsEndState checks if the current state is an end state.
// Returns true if the current state is one of the end states, false otherwise.
//
// Example:
//
//	if Machine.IsEndState() {
//	    // handle end state
//	}
func (m *Machine) IsEndState() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.endStates == nil {
		m.endStates = make(map[uint64]State)
		return false
	}

	curr, _ := m.curr.Load().(State)
	_, ok := m.endStates[curr.Id()]

	return ok
}

// AddTransition adds a transition to the Machine.
// If the Machine has no transitions, the from state of the transition is set as the start state.
// Panics if the transition does not have both from and to states.
//
// Example:
//
//	t := s1.When("condition", conditionFunc).Then(s2)
//	Machine.AddTransition(t)
func (m *Machine) AddTransition(t Transition) {
	m.mu.Lock()
	defer m.mu.Unlock()

	from := t.From()
	if from == nil || t.To() == nil {
		panic("transition must have a FROM and TO state")
	}

	if m.transitions == nil {
		m.transitions = make(map[uint64][]Transition)
		m.start.Store(from)
	}

	id := from.Id()
	m.transitions[id] = append(m.transitions[id], t)
}

// Validate validates the Machine configuration.
// It checks that:
// - The Machine has a start state
// - All transitions have from and to states
// - All state names are unique
//
// Returns an error if any of these conditions are not met.
//
// Example:
//
//	if err := Machine.Validate(); err != nil {
//	    // handle error
//	}
func (m *Machine) Validate() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sm := make(map[uint64]State)

	start, _ := m.start.Load().(State)
	if start == nil {
		return errors.New("no start state set")
	}
	var stateNames []string
	for _, tt := range m.transitions {
		for _, t := range tt {
			if t.From() == nil {
				return fmt.Errorf("transition '%s' has no from state", t.Description())
			}
			if t.To() == nil {
				return fmt.Errorf("transition '%s' has no to state", t.Description())
			}
			sm[t.From().Id()] = t.From()
			sm[t.To().Id()] = t.To()
			stateNames = append(stateNames, t.From().Name(), t.To().Name())
		}
	}

	stateNames = slice.String(stateNames).Unique()

	if len(stateNames) != len(sm) {
		return errors.New("invalid: all state names must be unique")
	}

	return nil
}

// Current returns the current state of the Machine.
// If the current state is not set, it returns the start state.
//
// Example:
//
//	currentState := Machine.Current()
//	fmt.Println("Current state:", currentState.Name())
func (m *Machine) Current() State {
	m.mu.RLock()
	defer m.mu.RUnlock()

	curr, _ := m.curr.Load().(State)
	if curr == nil {
		curr, _ = m.start.Load().(State)
	}
	return curr
}

// Update updates the Machine state based on the provided value.
// It evaluates all transitions from the current state and transitions to the first one
// whose condition evaluates to true.
//
// Returns:
// - bool: true if the state changed, false otherwise
// - error: any error that occurred during the update
//
// The update respects context cancellation.
//
// Example:
//
//	changed, err := Machine.Update(ctx, "some-value")
//	if err != nil {
//	    // handle error
//	}
//	if changed {
//	    fmt.Println("State changed to:", Machine.Current().Name())
//	}
func (m *Machine) Update(ctx context.Context, value interface{}) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	curr, _ := m.curr.Load().(State)
	if curr == nil {
		curr, _ = m.start.Load().(State)
		if curr == nil {
			return false, errors.New("machine has no start state")
		}
	}
	transitions, ok := m.transitions[curr.Id()]
	if !ok {
		return false, nil
	}

	for _, t := range transitions {
		success, err := t.Go(ctx, value)
		if err != nil {
			return false, err
		}

		if success {
			to := t.To()
			if to != nil {
				m.curr.Store(to)
			}

			return true, nil
		}
	}

	return false, nil
}
