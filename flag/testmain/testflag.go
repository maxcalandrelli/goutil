package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/maxcalandrelli/goutil/flag"
)

var (
	cmdopts struct {
		testf1    bool
		testf2    string
		opt_a_f1  int
		opt_a_f2  string
		opt_b_f1  string
		opt_b1_f1 int
	}
)

func testargs(args []string) {
	err := gu_flag.MainSet.Parse(args)
	if err != nil {
		fmt.Println("error: ", err.Error())
		return
	}
	fmt.Printf("'%s' parsed ok, cmdopts:%#v\n", strings.Join(args, " "), cmdopts)
}

func test(args string) {
	testargs(strings.Split(args, " "))
}

func main() {
	gu_flag.MainSet.Flags.BoolVar(&cmdopts.testf1, "i-meant-it", false, "global boolean")
	gu_flag.MainSet.Flags.StringVar(&cmdopts.testf2, "how", "gbl1", "global string")
	aopts := gu_flag.NewFlagSet("shoes", "set shoes size and description")
	aopts.Flags.IntVar(&cmdopts.opt_a_f1, "size", 38, "shoe size")
	aopts.Flags.StringVar(&cmdopts.opt_a_f2, "type", "boot", "shoe type")
	bopts := gu_flag.NewFlagSet("hat", "set hat characteristics")
	bopts.Flags.StringVar(&cmdopts.opt_b_f1, "style", "unknown", "hat style")
	b1opts := bopts.NewFlagSet("color", "set hat color")
	b1opts.Flags.IntVar(&cmdopts.opt_b1_f1, "index", -1, "color index")
	gu_flag.MainSet.Flags.Init("", flag.ContinueOnError)
	test("-bah")
	test("-how=notsohard shoes -size=33")
	test("-how=notsohard shoe -size=33")
	test("-how=notsohard shoes -size=33 --typ=comfort")
	test("-how=notsohard shoes -size=33 --type=comfort")
	test("hat")
	test("hat --style=classic")
	test("hat -style=havana color -index=4")
	testargs(os.Args[1:])
}
