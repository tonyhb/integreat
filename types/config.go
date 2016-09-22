package types

type Configuration struct {
	Base     Base
	Modules  []string
	Config   ModuleConfig
	Setup    []Test
	Tests    []Test
	Teardown []Test
}

type ModuleConfig map[string]map[string]interface{}

type Base struct {
	Version int
	Seed    int64
}

type Test struct {
	Id string

	// Name represents the human-readable name of the test
	Name string

	// Command describes the suite and function to call, in the format of
	// `Suite::FunctionName`
	Command string

	// Args is a map of arguments passed to the test command
	Args TestArgs

	// Repeat represents how many times this test will be repeated in sequence.
	// The default is 1.
	Repeat int
}
