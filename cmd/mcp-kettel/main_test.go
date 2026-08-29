package main

import (
	"strings"
	"testing"
)

func TestRunRequiresLabeledInput(t *testing.T) {
	for _, args := range [][]string{nil, {"/tmp/fastapi-project"}} {
		err := run(args)
		if err == nil || !strings.Contains(err.Error(), "--input DIR") {
			t.Fatalf("run(%q) error = %v", args, err)
		}
	}
}
