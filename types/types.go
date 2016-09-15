package types

// TestArgs is a map of interfaces representing arguments for a test command
// defined in the YAML file
type TestArgs map[string]interface{}

type TestResult map[string]interface{}

// TestCommand is the function signature of valid test commands callable from
// within a YAML file
type TestCommand func(TestArgs) (TestResult, error)

// Module is an interface representing a registerable suite of test commands
// for a specific product.
type Module interface {
	GetCommand(string) (TestCommand, error)
}

// ModuleCreator is a function which returns a concrete Module or an error
// given a map of config options.
//
// These config options should be returned from the specific config key of
// the main YAML config file.
type ModuleCreator func(map[string]interface{}) (Module, error)
