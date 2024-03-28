package pollution

import (
	"fmt"
	"strings"
	"testing"
)

func TestFaker(t *testing.T) {
	f := newFakerStrategy(map[string]interface{}{
		"position":  Beginning,
		"min_words": 2,
		"max_words": 3,
	})

	transformed, indicators := f.Apply("Some random text", "", "")
	fmt.Printf("Transformed: '%s' Indicators:%s\n", transformed, strings.Join(indicators, ","))
}
