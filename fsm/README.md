# FSM (Finite State Machine)

FSM is a flexible, thread-safe implementation of finite state machines in Go. It provides a clean API for modeling complex state transitions with support for context-aware operations.

## Features

- **Flexible State Definition**: Create states with unique names and IDs
- **Conditional Transitions**: Define transitions with custom conditions using Go functions
- **Start and End States**: Set specific start and end states for your state machine
- **Validation**: Validate your state machine configuration before use
- **Context-Aware**: All operations respect context cancellation
- **Thread-Safe**: All operations are thread-safe and can be used in concurrent environments
- **Minimal Dependencies**: Only depends on a single external package for slice operations

## Installation

```bash
go get github.com/twlvprscs/state/fsm
```

## Basic Usage

```go
package main

import (
	"context"
	"fmt"

	"github.com/twlvprscs/state/fsm"
)

func main() {
	// Create states
	s1 := fsm.NewState("STATE1")
	s2 := fsm.NewState("STATE2")
	s3 := fsm.NewState("STATE3")

	// Define transitions
	t1 := s1.When("value == 'a'", func(ctx context.Context, v interface{}) (bool, error) {
		if s, ok := v.(string); ok {
			return s == "a", nil
		}
		return false, nil
	}).Then(s2)

	t2 := s2.When("value == 'b'", func(ctx context.Context, v interface{}) (bool, error) {
		if s, ok := v.(string); ok {
			return s == "b", nil
		}
		return false, nil
	}).Then(s3)

	// Create machine with transitions
	machine := fsm.NewMachine(fsm.WithTransitions(t1, t2))

	// Set end states
	if err := machine.SetEndStates("STATE3"); err != nil {
		panic(err)
	}

	// Use the machine
	ctx := context.Background()
	
	// Transition to STATE2
	changed, err := machine.Update(ctx, "a")
	if err != nil {
		panic(err)
	}
	fmt.Printf("State changed: %v, Current state: %s\n", changed, machine.Current().Name())

	// Transition to STATE3
	changed, err = machine.Update(ctx, "b")
	if err != nil {
		panic(err)
	}
	fmt.Printf("State changed: %v, Current state: %s\n", changed, machine.Current().Name())

	// Check if we're in an end state
	if machine.IsEndState() {
		fmt.Println("Reached end state:", machine.Current().Name())
	}
}
```

## API Documentation

### Types

#### `State` Interface

The interface for state objects in the finite state machine.

```go
type State interface {
	Identifier
	Name() string
	When(string, TriggerFunc) Transition
}
```

#### `Transition` Interface

The interface for transitions between states.

```go
type Transition interface {
	Identifier
	Description() string
	From() State
	To() State
	Then(State) Transition
	Go(context.Context, interface{}) (bool, error)
}
```

#### `TriggerFunc` Type

Function type for transition conditions.

```go
type TriggerFunc func(context.Context, interface{}) (bool, error)
```

### Functions

#### `NewState`

```go
func NewState(name string, options ...StateOption) State
```

Creates a new state with the specified name and options.

#### `NewMachine`

```go
func NewMachine(opts ...Option) *machine
```

Creates a new finite state machine with the specified options.

#### Options

```go
// Adds transitions to the machine
func WithTransitions(transitions ...Transition) Option
```

### Methods

#### Machine Methods

```go
// Sets the start state by name
func (m *machine) SetStart(name string) error

// Resets the machine to its start state
func (m *machine) Reset() error

// Sets the end states by name
func (m *machine) SetEndStates(names ...string) error

// Checks if the current state is an end state
func (m *machine) IsEndState() bool

// Adds a transition to the machine
func (m *machine) AddTransition(t Transition)

// Validates the machine configuration
func (m *machine) Validate() error

// Returns the current state
func (m *machine) Current() State

// Updates the machine state based on the provided value
func (m *machine) Update(ctx context.Context, value interface{}) (bool, error)
```

#### State Methods

```go
// Returns the state name
func (s machineState) Name() string

// Creates a new transition from this state with the specified condition
func (s machineState) When(desc string, f TriggerFunc) Transition

// Returns the state ID
func (s machineState) Id() uint64
```

#### Transition Methods

```go
// Returns the transition ID
func (e *edge) Id() uint64

// Returns the transition description
func (e *edge) Description() string

// Returns the source state
func (e *edge) From() State

// Returns the destination state
func (e *edge) To() State

// Sets the destination state
func (e *edge) Then(s State) Transition

// Evaluates the transition condition
func (e *edge) Go(ctx context.Context, v interface{}) (bool, error)
```

## Implementation Details

The FSM package uses a few key design patterns:

1. **Interface-Based Design**: The core types (State, Transition) are defined as interfaces, allowing for flexible implementations.

2. **Builder Pattern**: The `When().Then()` pattern for creating transitions provides a fluent, readable API.

3. **Thread Safety**: The implementation uses a combination of `sync.RWMutex` and `atomic.Value` to ensure thread safety.

4. **Context Awareness**: All operations that might take time accept a context parameter, allowing for cancellation and timeouts.

5. **Unique IDs**: Each state and transition has a unique ID generated using atomic operations, ensuring uniqueness even in concurrent scenarios.

## Advanced Usage

### Complex State Machine

```go
package main

import (
	"context"
	"fmt"

	"github.com/twlvprscs/state/fsm"
)

func main() {
	// Create states
	idle := fsm.NewState("IDLE")
	connecting := fsm.NewState("CONNECTING")
	connected := fsm.NewState("CONNECTED")
	authenticating := fsm.NewState("AUTHENTICATING")
	authenticated := fsm.NewState("AUTHENTICATED")
	error := fsm.NewState("ERROR")
	
	// Define transitions
	transitions := []fsm.Transition{
		// From IDLE
		idle.When("connect requested", func(ctx context.Context, v interface{}) (bool, error) {
			cmd, ok := v.(string)
			return ok && cmd == "connect", nil
		}).Then(connecting),
		
		// From CONNECTING
		connecting.When("connection established", func(ctx context.Context, v interface{}) (bool, error) {
			evt, ok := v.(string)
			return ok && evt == "connected", nil
		}).Then(connected),
		
		connecting.When("connection failed", func(ctx context.Context, v interface{}) (bool, error) {
			evt, ok := v.(string)
			return ok && evt == "error", nil
		}).Then(error),
		
		// From CONNECTED
		connected.When("auth requested", func(ctx context.Context, v interface{}) (bool, error) {
			cmd, ok := v.(string)
			return ok && cmd == "authenticate", nil
		}).Then(authenticating),
		
		// From AUTHENTICATING
		authenticating.When("auth succeeded", func(ctx context.Context, v interface{}) (bool, error) {
			evt, ok := v.(string)
			return ok && evt == "auth_success", nil
		}).Then(authenticated),
		
		authenticating.When("auth failed", func(ctx context.Context, v interface{}) (bool, error) {
			evt, ok := v.(string)
			return ok && evt == "auth_failure", nil
		}).Then(error),
		
		// From ERROR
		error.When("retry", func(ctx context.Context, v interface{}) (bool, error) {
			cmd, ok := v.(string)
			return ok && cmd == "retry", nil
		}).Then(idle),
		
		// From any state
		authenticated.When("disconnect", func(ctx context.Context, v interface{}) (bool, error) {
			cmd, ok := v.(string)
			return ok && cmd == "disconnect", nil
		}).Then(idle),
		
		connected.When("disconnect", func(ctx context.Context, v interface{}) (bool, error) {
			cmd, ok := v.(string)
			return ok && cmd == "disconnect", nil
		}).Then(idle),
	}
	
	// Create machine
	machine := fsm.NewMachine(fsm.WithTransitions(transitions...))
	
	// Set end states
	machine.SetEndStates("AUTHENTICATED", "ERROR")
	
	// Use the machine
	ctx := context.Background()
	
	// Simulate a successful flow
	events := []string{"connect", "connected", "authenticate", "auth_success"}
	for _, evt := range events {
		changed, _ := machine.Update(ctx, evt)
		fmt.Printf("Event: %s, Changed: %v, Current: %s\n", 
			evt, changed, machine.Current().Name())
	}
	
	// Check if we're in an end state
	if machine.IsEndState() {
		fmt.Println("Reached end state:", machine.Current().Name())
	}
}
```

### Visualizing State Machines

The FSM package includes a `Graph()` method (currently a placeholder) that will eventually allow you to visualize your state machine. In the meantime, you can use the PlantUML format shown in the test file to visualize your state machines.

Example PlantUML:

```
@startuml

S1: IDLE
S2: CONNECTING
S3: CONNECTED
S4: AUTHENTICATING
S5: AUTHENTICATED
S6: ERROR

[*] --> S1
S1 --> S2 : connect
S2 --> S3 : connected
S2 --> S6 : error
S3 --> S4 : authenticate
S4 --> S5 : auth_success
S4 --> S6 : auth_failure
S6 --> S1 : retry
S5 --> S1 : disconnect
S3 --> S1 : disconnect

@enduml
```

## Potential Use Cases

The FSM package is ideal for scenarios where you need to model complex state transitions:

1. **Workflow Management**: Model business processes with multiple steps and conditions
2. **Protocol Implementations**: Implement network protocols with well-defined state transitions
3. **UI State Management**: Manage the state of complex user interfaces
4. **Game Development**: Model game states and transitions
5. **Parser Implementations**: Build parsers for complex grammars
6. **IoT Device Control**: Model the behavior of IoT devices
7. **Transaction Processing**: Model the lifecycle of transactions
8. **Document Processing**: Track the state of documents in a workflow