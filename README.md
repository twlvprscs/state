# State

A Go library for state management, providing flexible and efficient tools for handling state in your applications.

## Overview

The `state` library offers two main packages:

1. **FSM (Finite State Machine)**: A general-purpose finite state machine implementation for modeling complex state transitions.
2. **Switchboard**: A high-performance, concurrent-safe mechanism for managing binary states (open/closed) and triggering events when those states change.

## Installation

```bash
# Install the entire library
go get github.com/twlvprscs/state

# Or install individual packages
go get github.com/twlvprscs/state/fsm
go get github.com/twlvprscs/state/switchboard
```

## Packages

### FSM

The FSM package provides a flexible implementation of finite state machines with the following features:

- Define states and transitions between them
- Set start and end states
- Validate machine configurations
- Context-aware operations
- Thread-safe implementation

#### Basic Usage

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
    t1 := s1.When("v == a", func(ctx context.Context, v interface{}) (bool, error) {
        return v == "a", nil
    }).Then(s2)
    
    t2 := s2.When("v == b", func(ctx context.Context, v interface{}) (bool, error) {
        return v == "b", nil
    }).Then(s3)
    
    // Create machine with transitions
    machine := fsm.NewMachine(fsm.WithTransitions(t1, t2))
    
    // Set end states
    machine.SetEndStates("STATE3")
    
    // Use the machine
    ctx := context.Background()
    machine.Update(ctx, "a") // Transition to STATE2
    machine.Update(ctx, "b") // Transition to STATE3
    
    // Check if we're in an end state
    if machine.IsEndState() {
        fmt.Println("Reached end state:", machine.Current().Name())
    }
}
```

### Switchboard

Switchboard is a high-performance, concurrent-safe mechanism for managing binary states (open/closed) and triggering events when those states change. It efficiently handles up to 4096 different states using bit manipulation.

#### Features

- **Efficient State Management**: Uses bit manipulation to efficiently track up to 4096 different states
- **Concurrent-Safe**: All operations are thread-safe and can be used in concurrent environments
- **Event-Driven**: Register handlers to be notified when states change
- **Context-Aware**: All operations respect context cancellation
- **Minimal Memory Footprint**: Uses only 512 bytes (64 uint64 words) to track all states

#### Basic Usage

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

For more detailed documentation and examples, please refer to the README files in each package:
- [FSM Package Documentation](fsm/README.md)
- [Switchboard Package Documentation](switchboard/README.md)

## Use Cases

The `state` library is suitable for a variety of applications:

- **Application State Management**: Track the state of various components in your application
- **Workflow Management**: Model complex workflows with multiple steps and conditions
- **Event-Driven Systems**: Build event-driven systems where actions are triggered by state changes
- **Feature Flags**: Implement feature flags that can be toggled at runtime
- **Circuit Breakers**: Implement circuit breakers for resilient systems
- **Permission Systems**: Track user permissions and access rights
- **Resource Management**: Monitor the availability of resources
- **Distributed Systems**: Coordinate state across distributed systems

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.