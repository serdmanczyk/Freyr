package envflags

import (
	"flag"
	"os"
	"reflect"
	"testing"
)

type testConfig struct {
	VarOne   string `flag:"varone" env:"TEST_VARONE"`
	VarTwo   string `flag:"vartwo" env:"TEST_VARTWO"`
	VarThree string `flag:"varthree" env:"TEST_VARTHREE"`
	VarFour  string `flag:"varfour" env:"TEST_VARFOUR"`
}

func TestFlags(t *testing.T) {
	var c testConfig

	os.Args = []string{"ignore", "--varone", "varone", "--vartwo", "vartwo"}
	os.Setenv("TEST_VARTHREE", "varthree")

	SetFlags(&c)

	flag.Parse()

	if !ConfigEmpty(&c) {
		t.Errorf("Config should be empty, flag four is not set")
	}

	e := testConfig{"varone", "vartwo", "varthree", ""}
	if !reflect.DeepEqual(c, e) {
		t.Errorf("Variables values incorrect: %v expected: %v", c, e)
	}
}

// Could run more tests but need to figure out how to reset 'flag' package.
// Not currently a priority.
