package utils

import (
	"fmt"
	"testing"
)

func TestRotateMap(t *testing.T) {
	m := map[string]int{
		"1": 1,
		"2": 2,
	}

	delete(m, "1")
	newMap := RotateMap(m)
	fmt.Printf("%#v\n", newMap)
}
