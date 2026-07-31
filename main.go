package main

import (
	"encoding/json"
	"fmt"
	"os"

	tc "github.com/NickEngland21/port-mortem-tinycolor-go"
)

type fixtureFile struct {
	Cases []fixture `json:"cases"`
}

type fixture struct {
	Input  any    `json:"input"`
	Valid  bool   `json:"valid"`
	Format any    `json:"format"`
	Hex8   string `json:"hex8"`
	String string `json:"string"`
}

type failure struct {
	Input    any     `json:"input"`
	Expected fixture `json:"expected"`
	Actual   result  `json:"actual"`
}

type result struct {
	Valid  bool   `json:"valid"`
	Format any    `json:"format"`
	Hex8   string `json:"hex8"`
	String string `json:"string"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: differential FIXTURES_JSON")
		os.Exit(2)
	}
	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var file fixtureFile
	if err := json.Unmarshal(b, &file); err != nil {
		panic(err)
	}
	failures := make([]failure, 0)
	for _, f := range file.Cases {
		c := tc.New(f.Input)
		format := any(c.GetFormat())
		if c.GetFormat() == "" {
			format = false
		}
		actual := result{Valid: c.IsValid(), Format: format, Hex8: c.ToHex8(), String: c.ToString()}
		expectedFormat := f.Format
		if expectedFormat == nil {
			expectedFormat = false
		}
		if actual.Valid != f.Valid || actual.Format != expectedFormat || actual.Hex8 != f.Hex8 || actual.String != f.String {
			failures = append(failures, failure{Input: f.Input, Expected: f, Actual: actual})
		}
	}
	out := map[string]any{"ok": len(failures) == 0, "cases": len(file.Cases), "failures": failures, "fixtures": os.Args[1]}
	enc, _ := json.Marshal(out)
	fmt.Println(string(enc))
	if len(failures) > 0 {
		os.Exit(1)
	}
}
