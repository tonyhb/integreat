package errors

import (
	"fmt"
)

var (
	// ErrModuleRegistered is returned when attempting to register a suite
	// of tests with a name that has already been registered; the namespace
	// is conflicting.
	ErrModuleRegistered = fmt.Errorf("module is already registered")
	// ErrModuleUnregistered is returned when a YAML file references
	// a suite of tests which has not yet been registered
	ErrModuleUnregistered = fmt.Errorf("module is not registered")
	ErrCommandNotFound    = fmt.Errorf("command not found")
)
