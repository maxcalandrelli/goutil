// Repeatable flags can handle simple cases of multiple values, like
// lists or sets
package gu_flag

import (
	"errors"
	"fmt"
	"strings"
)

var (
	DUPLICATE_VALUE error = errors.New("value already present")
	NOT_IN_SET      error = errors.New("value out of set")
)

type RepeatableArg interface {
	GetSeparator() string
	SetSeparator(string)
	IsDefault() bool
	setDefault(bool)
	getCanonicalizer() func(string) string
	setCanonicalizer(func(string) string)
	getInserter() func(string) error
	setInserter(func(string) error)
	GetValues() []string
	String() string
	reset()
}

type listOrSetArg struct {
	RepeatableArg
	separator     string
	isDefault     bool
	canonicalizer func(string) string
	inserter      func(string) error
	values        *[]string
}

type setArg struct {
	listOrSetArg
	values map[string]bool
}

func (losa *listOrSetArg) String() string {
	if v := losa.GetValues(); losa != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func (losa *listOrSetArg) GetSeparator() string {
	return losa.separator
}

func (losa *listOrSetArg) SetSeparator(s string) {
	losa.separator = s
}

func (losa *listOrSetArg) IsDefault() bool {
	return losa.isDefault
}

func (losa *listOrSetArg) setDefault(d bool) {
	losa.isDefault = d
}

func (losa *listOrSetArg) getCanonicalizer() func(string) string {
	return losa.canonicalizer
}

func (losa *listOrSetArg) setCanonicalizer(canonicalizer func(string) string) {
	losa.canonicalizer = canonicalizer
}

func (losa *listOrSetArg) getInserter() func(string) error {
	return losa.inserter
}

func (losa *listOrSetArg) setInserter(inserter func(string) error) {
	losa.inserter = inserter
}

func (losa *listOrSetArg) Get() interface{} {
	return losa.GetValues()
}

func (losa *listOrSetArg) Set(v string) error {
	if losa.IsDefault() {
		losa.reset()
		losa.setDefault(false)
	}
	for _, _v := range strings.Split(v, losa.GetSeparator()) {
		if err := losa.inserter(losa.canonicalizer(_v)); err != nil {
			return err
		}
	}
	return nil
}

func (losa *listOrSetArg) GetValues() []string {
	if losa.values == nil {
		return []string{}
	}
	return *(losa.values)
}

func (losa *listOrSetArg) reset() {
	*(losa.values) = []string{}
}

func newListArg(values *[]string, separator string, default_value []string, canonicalizer func(string) string, inserter func(string) error) *listOrSetArg {
	la := listOrSetArg{
		values:        values,
		separator:     separator,
		canonicalizer: canonicalizer,
		inserter:      inserter,
	}
	for _, _v := range default_value {
		la.Set(_v)
	}
	la.setDefault(true)
	return &la
}

func (sa *setArg) reset() {
	sa.values = map[string]bool{}
	sa.listOrSetArg.reset()
}

func (f *FlagSet) List(name string, value []string, usage string) RepeatableArg {
	return f.ListVar(new([]string), name, value, usage)
}

func (f *FlagSet) ListVar(p *[]string, name string, value []string, usage string) RepeatableArg {
	values := newListArg(p, ",", value, func(v string) string { return v }, func(v string) error { *p = append(*p, v); return nil })
	f.Flags.Var(values, name, usage)
	return values
}

func (f *FlagSet) Set(name string, value []string, usage string, ignoreCase bool) RepeatableArg {
	return f.SetVar(new([]string), name, value, usage, ignoreCase)
}

func (f *FlagSet) SetVar(p *[]string, name string, value []string, usage string, ignoreCase bool) RepeatableArg {
	vmap := map[string]bool{}
	sa := setArg{
		values: vmap,
		listOrSetArg: *newListArg(
			p,
			",",
			value,
			func(v string) string {
				if ignoreCase {
					return strings.ToLower(v)
				}
				return v
			},
			func(v string) error {
				if _, _present := vmap[v]; _present {
					return DUPLICATE_VALUE
				}
				vmap[v] = true
				*p = append(*p, v)
				return nil
			},
		),
	}
	for _, _v := range value {
		sa.Set(_v)
	}
	sa.setDefault(true)
	f.Flags.Var(&sa, name, usage)
	return &sa
}

func (f *FlagSet) ConstrainedSet(name string, value []string, allowed []string, usage string, ignoreCase bool) RepeatableArg {
	return f.ConstrainedSetVar(new([]string), name, value, allowed, usage, ignoreCase)
}

func (f *FlagSet) ConstrainedSetVar(p *[]string, name string, value []string, allowed []string, usage string, ignoreCase bool) RepeatableArg {
	sa := f.SetVar(p, name, value, usage, ignoreCase)
	validMap := map[string]bool{}
	for _, _v := range allowed {
		validMap[sa.getCanonicalizer()(_v)] = true
	}
	_i := sa.getInserter()
	sa.setInserter(func(v string) error {
		if _, _valid := validMap[v]; _valid {
			return _i(v)
		}
		return NOT_IN_SET
	})
	return sa
}
