package eval

import (
	"path/filepath"
	"testing"

	"github.com/andrewpillar/req/syntax"
	"github.com/andrewpillar/req/token"
)

func errh(t *testing.T) func(token.Pos, string) {
	return func(pos token.Pos, msg string) {
		t.Errorf("%s - %s\n", pos, msg)
	}
}

func Test_EvalVarDecl(t *testing.T) {
	nn, err := syntax.ParseFile(filepath.Join("testdata", "vardecl.req"), errh(t))

	if err != nil {
		t.Fatal(err)
	}

	var (
		e Evaluator
		c Context
	)

	for _, n := range nn {
		if _, err := e.Eval(&c, n); err != nil {
			t.Errorf("%s\n", err)
		}
	}

	tests := []struct {
		varname  string
		expected Type
	}{
		{"String", String},
		{"Number", Int},
		{"Bool", Bool},
		{"Array", Array},
		{"Hash", Hash},
	}

	for i, test := range tests {
		obj, err := c.Get(test.varname)

		if err != nil {
			t.Errorf("tests[%d] - %s\n", i, err)
			continue
		}

		if typ := obj.Type(); typ != test.expected {
			t.Errorf("tests[%d] - unexpected type for variable %q, expected=%q, got=%q\n", i, test.varname, test.expected, typ)
		}
	}
}

func Test_EvalRef(t *testing.T) {
	nn, err := syntax.ParseFile(filepath.Join("testdata", "refexpr.req"), errh(t))

	if err != nil {
		t.Fatal(err)
	}

	var e Evaluator

	e.AddCmd(PrintCmd)

	if err := e.Run(nn); err != nil {
		t.Fatal(err)
	}
}

func Test_EvalInterpolate(t *testing.T) {
	nn, err := syntax.ParseFile(filepath.Join("testdata", "vardecl.req"), errh(t))

	if err != nil {
		t.Fatal(err)
	}

	var (
		e Evaluator
		c Context
	)

	for _, n := range nn {
		if _, err := e.Eval(&c, n); err != nil {
			t.Errorf("%s\n", err)
		}
	}

	tests := []struct {
		input    string
		expected string
	}{
		{`{$String}`, "string"},
		{`{$Number}`, "10"},
		{`{$Bool}`, "true"},
		{`{$Array[0]}`, "1"},
		{`{$Array[2]}`, "3"},
		{`{$Array[3]}`, "4"},
		{`{$Hash["String"]}`, "string"},
		{`{$Hash["Array"][0]}`, "1"},
		{`{$Hash["Child"]["Array"][2]}`, "three"},
	}

	for i, test := range tests {
		obj, err := e.interpolate(&c, test.input)

		if err != nil {
			t.Errorf("tests[%d] - failed to interpolate string: %s\n", i, err)
			continue
		}

		s, ok := obj.(stringObj)

		if !ok {
			t.Fatalf("tests[%d] - Eval.interpolate did not return a stringObj", i)
		}

		if s.value != test.expected {
			t.Errorf("tests[%d] - unexpected output for %q, expected=%q, got=%q\n", i, test.input, test.expected, s.value)
		}
	}
}
