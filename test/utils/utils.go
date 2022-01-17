package utils

import (
	"fmt"
	"testing"
)

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func Assert(t *testing.T, expected, actual interface{}, msg string) bool {
	if expected != actual {
		t.Error(msg, fmt.Sprintf("expected : %v, got : %v", expected, actual))
		return false
	}
	return true
}
