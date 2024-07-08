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
}

type Command struct {
	HasArg  bool
	Action  func(string)
	Options []string
}

func Constructor(debug bool) *ArgumentParser {
	ap := &ArgumentParser{
		flagSet:  flag.NewFlagSet("argumentparser", flag.ExitOnError),
		commands: make(map[string]*Command),
	}
	return ap
}

func (ap *ArgumentParser) AddFunction(name, shorthand string, required, hasArg bool, description string, action interface{}) {
	wrappedAction := wrapAction(action)
	if hasArg {
		ap.flagSet.String(shorthand, "", description)
	} else {
		ap.flagSet.Bool(shorthand, false, description)
	}
	ap.commands[shorthand] = &Command{HasArg: hasArg, Action: wrappedAction}
}

func (ap *ArgumentParser) AddFunctionOptions(name, shorthand string, required, hasArg bool, description string, action interface{}, options []string) {
	ap.AddFunction(name, shorthand, required, hasArg, description, action)
	ap.commands[shorthand].Options = options
}

func (ap *ArgumentParser) Parse() {
	ap.flagSet.Parse(os.Args[1:])
	ap.flagSet.Visit(func(f *flag.Flag) {
		if cmd, found := ap.commands[f.Name]; found {
			if cmd.HasArg {
				cmd.Action(f.Value.String())
			} else {
				cmd.Action("")
			}
		}
	})
}

func wrapAction(action interface{}) func(string) {
	v := reflect.ValueOf(action)
	t := v.Type()

	fmt.Printf("Wrapping action of type: %T\n", action) // Debug print

	switch t.Kind() {
	case reflect.Func:
		switch t.NumIn() {
		case 0:
			if t.NumOut() == 0 {
				return func(string) { v.Call(nil) }
			}
			if t.NumOut() == 1 {
				return func(string) { v.Call(nil) }
			}
		case 1:
			if t.In(0).Kind() == reflect.String {
				if t.NumOut() == 0 {
					return func(arg string) { v.Call([]reflect.Value{reflect.ValueOf(arg)}) }
				}
				if t.NumOut() == 1 {
					return func(arg string) { v.Call([]reflect.Value{reflect.ValueOf(arg)}) }
				}
			}
			if t.In(0).Kind() == reflect.Bool {
				return func(arg string) { v.Call([]reflect.Value{reflect.ValueOf(arg == "true")}) }
			}
			if t.In(0).Kind() == reflect.Slice && t.In(0).Elem().Kind() == reflect.String {
				return func(arg string) { v.Call([]reflect.Value{reflect.ValueOf([]string{arg})}) }
			}
		}
	}

	panic(fmt.Sprintf("unsupported action type: %T", action))
}
