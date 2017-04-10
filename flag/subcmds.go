// Keyword arguments can open sub-grammars
// having their own flags and/or keywords, that in turn can open
// other subgrammars
package gu_flag

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

type (
	// FlagSet represents the subgrammar introduced by a given keyword
	FlagSet struct {
		subCommand     string
		usage          string
		Flags          *flag.FlagSet
		subSets        map[string]*FlagSet
		parent         *FlagSet
		NonKeywordArgs bool
		HookFunc       func(self *FlagSet) error
	}
)

var (
	ForceEqualAlways         = true
	MainSet          FlagSet = FlagSet{Flags: flag.CommandLine, subSets: map[string]*FlagSet{}}
)

func (f *FlagSet) PrintDefaults() {
	if f.parent != nil {
		fmt.Fprintf(os.Stderr, "\n  %s: %s\n\n", f.subCommand, f.usage)
	}
	f.Flags.PrintDefaults()
	for _, sub := range f.subSets {
		sub.PrintDefaults()
	}
}

func (f *FlagSet) Keywords() []string {
	r := []string{}
	for k, _ := range f.subSets {
		r = append(r, k)
	}
	return r
}

func (f *FlagSet) Name() string {
	return f.subCommand
}

func (f *FlagSet) Parent() *FlagSet {
	return f.parent
}

func (f *FlagSet) Usage() string {
	return f.usage
}

func (f *FlagSet) Parse(args []string) error {
	if ForceEqualAlways {
		flags := []string{}
		for len(args) > 0 && args[0][0] == '-' {
			flags = append(flags, args[0])
			args = args[1:]
		}
		if len(flags) > 0 {
			if err := f.Flags.Parse(flags); err != nil {
				return err
			}
		}
	} else {
		if err := f.Flags.Parse(args); err != nil {
			return err
		}
		args = f.Flags.Args()
	}
	switch {
	case len(args) == 0 && len(f.subSets) == 0:
		return f.HookFunc(f)
	case len(args) == 0 && len(f.subSets) != 0:
		return errors.New(fmt.Sprintf("missing keyword (try one of: %s)", strings.Join(f.Keywords(), ",")))
	case len(args) != 0 && len(f.subSets) == 0:
		if f.NonKeywordArgs {
			return f.HookFunc(f)
		} else {
			return errors.New(fmt.Sprintf("unexpected: %s", args[0]))
		}
	}
	if s, ok := f.subSets[args[0]]; ok {
		return s.Parse(args[1:])
	} else {
		return errors.New(fmt.Sprintf("unknown keyword: %s", args[0]))
	}
}

func (f *FlagSet) NewFlagSet(name, usage string) *FlagSet {
	r := new(FlagSet)
	r.subSets = make(map[string]*FlagSet)
	r.subCommand = name
	r.usage = usage
	r.Flags = flag.NewFlagSet(name, flag.ContinueOnError)
	r.parent = f
	if r.parent == nil {
		r.parent = &MainSet
	}
	r.parent.subSets[name] = r
	r.HookFunc = func(*FlagSet) error { return nil }
	return r
}

func NewFlagSet(name, usage string) *FlagSet {
	return (*FlagSet)(nil).NewFlagSet(name, usage)
}
