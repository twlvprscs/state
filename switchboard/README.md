# Switchboard

Switchboard is a high-performance, concurrent-safe mechanism for managing binary states (open/closed) and triggering events when those states change. It efficiently handles up to 4096 different states using bit manipulation.

## Features

- **Efficient State Management**: Uses bit manipulation to efficiently track up to 4096 different states
- **Concurrent-Safe**: All operations are thread-safe and can be used in concurrent environments
- **Event-Driven**: Register handlers to be notified when states change
- **Context-Aware**: All operations respect context cancellation
- **Minimal Memory Footprint**: Uses only 512 bytes (64 uint64 words) to track all states

## Installation

```bash
go get github.com/twlvprscs/state/switchboard
```

## Usage

### Basic Usage

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/twlvprscs/state/switchboard"
)

func main() {
	// Create a new switchboard
	sb := switchboard.New()

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the state machine
	sb.Run(ctx)

	// Define some state indices
	const (
		SystemReady = iota
		DatabaseConnected
		UserAuthenticated
		DataLoaded
	)

	// Close (activate) some states
	sb.Close(ctx, SystemReady, DatabaseConnected)

	// Open (deactivate) a state
	sb.Open(ctx, DatabaseConnected)

	// Toggle a state
	sb.Toggle(ctx, UserAuthenticated)

	// Print the current state
	fmt.Printf("Current state: %#v\n", sb)

	// Wait for a moment to allow handlers to process
	time.Sleep(100 * time.Millisecond)
}
```

### With State Change Handlers

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/twlvprscs/state/switchboard"
)

func main() {
	// Define some state indices
	const (
		SystemReady = iota
		DatabaseConnected
		UserAuthenticated
		DataLoaded
	)

	// Create a new state machine with handlers
	machine := switchboard.New(
		// Handler for all state changes
		switchboard.WithDefaultChangeHandler(func(ctx context.Context, idx uint, closed bool) {
			stateStr := "opened"
			if closed {
				stateStr = "closed"
			}
			fmt.Printf("State %d was %s\n", idx, stateStr)
		}),
		
		// Handler for a specific state
		switchboard.WithSingleStateChangeHandler(func(ctx context.Context, closed bool) {
			if closed {
				fmt.Println("Database connected!")
			} else {
				fmt.Println("Database disconnected!")
			}
		}, DatabaseConnected),
		
		// Initialize with all states closed
		switchboard.WithAllStatesClosed(),
	)

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the state machine
	machine.Run(ctx)

	// Manipulate states
	machine.Open(ctx, DatabaseConnected)
	machine.Close(ctx, DatabaseConnected)
	machine.Toggle(ctx, UserAuthenticated)

	// Wait for handlers to process
	time.Sleep(100 * time.Millisecond)
}
```

## Potential Use Cases

Switchboard is ideal for scenarios where you need to track multiple binary states and react to changes:

1. **Application State Management**: Track the state of various components in your application
2. **Feature Flags**: Implement feature flags that can be toggled at runtime
3. **Dependency Tracking**: Monitor the availability of external dependencies
4. **Event-Driven Systems**: Build event-driven systems where actions are triggered by state changes
5. **Workflow Management**: Track the progress of complex workflows with multiple steps
6. **Circuit Breakers**: Implement circuit breakers for resilient systems
7. **Permission Systems**: Track user permissions and access rights
8. **Resource Management**: Monitor the availability of resources
9. **Distributed Systems**: Coordinate state across distributed systems
10. **IoT Applications**: Track the state of connected devices

## API Documentation

### Types

#### `Switch` Interface

The main interface for interacting with the state machine.

```go
type Switch interface {
	fmt.GoStringer
	Open(ctx context.Context, conditions ...uint)
	Close(ctx context.Context, conditions ...uint)
	Toggle(ctx context.Context, conditions ...uint)
	Run(ctx context.Context)
}
```

#### `S` Struct

The main implementation of the `Switch` interface.

```go
type S struct {
	// Contains unexported fields
}
```

#### Handler Types

```go
// Handles state changes for any condition
type ChangeHandler func(ctx context.Context, idx uint, state bool)

// Handles state changes for a specific condition
type SingleStateChangeHandler func(ctx context.Context, state bool)
```

### Functions

#### `NewMachine`

```go
func NewMachine(opts ...Option) *S
```

Creates a new state machine with the specified options.

#### Options

```go
// Sets a handler for all state changes
func WithDefaultChangeHandler(handler ChangeHandler) Option

// Sets a handler for a specific state
func WithSingleStateChangeHandler(handler SingleStateChangeHandler, condition uint) Option

// Initializes the state machine with all states closed
func WithAllStatesClosed() Option
```

### Methods

#### `Run`

```go
func (s *S) Run(ctx context.Context)
```

Starts the state machine, enabling it to process state changes and notify handlers.

#### `Close`

```go
func (s *S) Close(ctx context.Context, conditions ...uint)
```

Sets the specified conditions to the closed state.

#### `Open`

```go
func (s *S) Open(ctx context.Context, conditions ...uint)
```

Sets the specified conditions to the open state.

#### `Toggle`

```go
func (s *S) Toggle(ctx context.Context, conditions ...uint)
```

Switches the state of the specified conditions.

#### `GoString`

```go
func (s *S) GoString() string
```

Returns a string representation of the state machine's current state.

## Implementation Details

Internally, Switchboard uses a bit manipulation approach to efficiently track states:

- States are represented as bits in an array of uint64 values
- Each uint64 can track 64 states, and with 64 uint64s, a total of 4096 states can be tracked
- A state can be either "open" (0) or "closed" (1)
- Operations are performed using bitwise operations for maximum efficiency
- Thread safety is ensured using a channel-based locking mechanism
- Event notification is handled through a dedicated goroutine