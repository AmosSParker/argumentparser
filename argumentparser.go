package argumentparser

import (
	"flag"
	"fmt"
	"os"
	"reflect"
)

type ArgumentParser struct {
	flagSet  *flag.FlagSet
	commands map[string]*Command
	results  map[string]string
}

type Command struct {
	HasArg  bool
	Action  func(string)
	Options []string
}

// Constructor initializes a new ArgumentParser
func Constructor(debug bool) *ArgumentParser {
	return &ArgumentParser{
		flagSet:  flag.NewFlagSet("argumentparser", flag.ExitOnError),
		commands: make(map[string]*Command),
		results:  make(map[string]string),
	}
}

// AddFunction adds a command with an associated action
func (ap *ArgumentParser) AddFunction(name, shorthand string, required, hasArg bool, description string, action interface{}) {
	wrappedAction := wrapAction(action)
	if hasArg {
		ap.flagSet.String(shorthand, "", description)
	} else {
		ap.flagSet.Bool(shorthand, false, description)
	}
	ap.commands[shorthand] = &Command{HasArg: hasArg, Action: wrappedAction}
}

// AddFunctionOptions adds a command with options and an associated action
func (ap *ArgumentParser) AddFunctionOptions(name, shorthand string, required, hasArg bool, description string, action interface{}, options []string) {
	ap.AddFunction(name, shorthand, required, hasArg, description, action)
	ap.commands[shorthand].Options = options
}

// Add adds a command without an associated action
func (ap *ArgumentParser) Add(name, shorthand string, required, hasArg bool, description string) {
	ap.AddFunction(name, shorthand, required, hasArg, description, func(string) {})
}

// Parse processes the command-line arguments
func (ap *ArgumentParser) Parse() map[string]string {
	ap.flagSet.Parse(os.Args[1:])
	ap.flagSet.Visit(func(f *flag.Flag) {
		if cmd, found := ap.commands[f.Name]; found {
			if cmd.HasArg {
				ap.results[f.Name] = f.Value.String()
				cmd.Action(f.Value.String())
			} else {
				ap.results[f.Name] = "true"
				cmd.Action("")
			}
		}
	})
	return ap.results
}

// wrapAction wraps an action function to match the required interface
func wrapAction(action interface{}) func(string) {
	if action == nil {
		panic("Error: Action cannot be nil")
	}

	v := reflect.ValueOf(action)
	t := v.Type()

	// Debug print to understand action type
	fmt.Printf("Wrapping action of type: %T\n", action)

	// Validate that the action is a function
	if t.Kind() != reflect.Func {
		panic(fmt.Sprintf("unsupported action type: %T (must be a function)", action))
	}

	switch t.NumIn() {
	case 0: // No arguments
		if t.NumOut() == 0 {
			return func(string) { v.Call(nil) }
		} else if t.NumOut() == 1 {
			return func(string) { v.Call(nil) }
		}
	case 1: // Single argument
		if t.In(0).Kind() == reflect.String {
			if t.NumOut() == 0 {
				return func(arg string) { v.Call([]reflect.Value{reflect.ValueOf(arg)}) }
			} else if t.NumOut() == 1 {
				return func(arg string) { v.Call([]reflect.Value{reflect.ValueOf(arg)}) }
			}
		} else if t.In(0).Kind() == reflect.Bool {
			return func(arg string) { v.Call([]reflect.Value{reflect.ValueOf(arg == "true")}) }
		} else if t.In(0).Kind() == reflect.Slice && t.In(0).Elem().Kind() == reflect.String {
			return func(arg string) { v.Call([]reflect.Value{reflect.ValueOf([]string{arg})}) }
		}
	default:
		panic(fmt.Sprintf("unsupported action function signature: %T", action))
	}

	panic(fmt.Sprintf("unsupported action type: %T", action))
}

