// Package switchboard provides a high-performance, concurrent-safe mechanism for
// managing binary states (open/closed) and triggering events when those states change.
// It efficiently handles up to 4096 different states using bit manipulation.
package switchboard

import (
	"context"
	"fmt"
)

// ChangeHandler is a function that handles state changes for any condition.
// It receives the context, the index of the changed state, and whether it was closed (true) or opened (false).
type ChangeHandler func(ctx context.Context, idx uint, state bool)

// SingleStateChangeHandler is a function that handles state changes for a specific condition.
// It receives the context and whether the state was closed (true) or opened (false).
type SingleStateChangeHandler func(ctx context.Context, state bool)

// Switch defines the interface for a state machine that can manage binary states.
type Switch interface {
	fmt.GoStringer
	// Open sets the specified conditions to the open state.
	// If a condition changes state, registered handlers will be notified.
	Open(ctx context.Context, conditions ...uint)
	// Close sets the specified conditions to the closed state.
	// If a condition changes state, registered handlers will be notified.
	Close(ctx context.Context, conditions ...uint)
	// Toggle switches the state of the specified conditions.
	// If a condition changes state, registered handlers will be notified.
	Toggle(ctx context.Context, conditions ...uint)
	// Run starts the state machine, enabling it to process state changes and notify handlers.
	// This method should be called before using the state machine.
	Run(ctx context.Context)
}

// S is the main implementation of the Switch interface.
// It manages a set of binary states and notifies registered handlers when states change.
type S struct {
	delegate      *delegate
	defaultChange func(context.Context, uint, bool)
	changeMap     map[uint]func(context.Context, bool)
}

// Ensure S implements the Switch interface
var _ Switch = (*S)(nil)

// Option is a function that configures an S instance.
type Option func(*S)

// WithDefaultChangeHandler sets a handler that will be called for all state changes
// that don't have a specific handler registered.
func WithDefaultChangeHandler(handler ChangeHandler) Option {
	return func(s *S) {
		s.defaultChange = handler
	}
}

// WithSingleStateChangeHandler registers a handler for a specific condition.
// This handler will be called when the specified condition changes state.
func WithSingleStateChangeHandler(handler SingleStateChangeHandler, condition uint) Option {
	return func(s *S) {
		s.changeMap[condition] = handler
	}
}

// WithAllStatesClosed initializes the switchboard with all states set to closed.
// By default, all states are initialized as open.
func WithAllStatesClosed() Option {
	return func(s *S) {
		defer s.delegate.lock().unlock()
		s.delegate.reg = registerWithAllClosed()
	}
}

// New creates a new switchboard with the specified options.
// By default, all states are initialized as open and no handlers are registered.
func New(opts ...Option) *S {
	s := S{
		delegate:      newDelegate(),
		defaultChange: func(context.Context, uint, bool) {},
		changeMap:     make(map[uint]func(context.Context, bool)),
	}

	for _, f := range opts {
		f(&s)
	}

	return &s
}

// Run starts the state machine, enabling it to process state changes and notify handlers.
// It launches a goroutine that listens for state changes and calls the appropriate handlers.
// The goroutine will run until the provided context is canceled.
func (s *S) Run(ctx context.Context) {
	go func(ctx context.Context, s *S) {
		for {
			select {
			case <-ctx.Done():
				return
			case c := <-s.delegate.changeChan:
				if f, ok := s.changeMap[c.state]; ok {
					f(c.ctx, c.closed)
					continue
				}
				s.defaultChange(c.ctx, c.state, c.closed)
			}
		}
	}(ctx, s)
}

// Close sets the specified conditions to the closed state.
// If a condition changes state, registered handlers will be notified.
// This method is safe for concurrent use.
func (s *S) Close(ctx context.Context, conditions ...uint) {
	s.delegate.close(ctx, conditions...)
}

// Open sets the specified conditions to the open state.
// If a condition changes state, registered handlers will be notified.
// This method is safe for concurrent use.
func (s *S) Open(ctx context.Context, conditions ...uint) {
	s.delegate.open(ctx, conditions...)
}

// Toggle switches the state of the specified conditions.
// If a condition is open, it will be closed, and vice versa.
// Registered handlers will be notified of any state changes.
// This method is safe for concurrent use.
func (s *S) Toggle(ctx context.Context, conditions ...uint) {
	s.delegate.toggle(ctx, conditions...)
}

// GoString returns a string representation of the state machine's current state.
// This implements the fmt.GoStringer interface.
func (s *S) GoString() string {
	return s.delegate.stringVal()
}
