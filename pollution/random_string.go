package pollution

import (
	"fmt"
	"github.com/melardev/discord-message-protect/utils"
	"math/rand"
	"strings"
)

type RandomStringStrategy struct {
	Position IndicatorPosition
	MinWords int
	MaxWords int

	MinWordLen int
	MaxWordLen int
}

func (f *RandomStringStrategy) GetName() string {
	return RandomString
}

func newRandomStringStrategy(dto map[string]interface{}) *RandomStringStrategy {
	minWords := 0
	if val, ok := dto["min_words"].(float64); ok {
		minWords = int(val)
	} else if val2, ok := dto["min_words"].(int); ok {
		minWords = val2
	}

	maxWords := 0
	if val, ok := dto["max_words"].(float64); ok {
		maxWords = int(val)
	} else if val2, ok := dto["max_words"].(int); ok {
		maxWords = val2
	}

	var position IndicatorPosition
	if pos, ok := dto["position"].(int); ok {
		position = IndicatorPosition(pos)
	} else if pos2, ok2 := dto["position"].(string); ok2 {
		position = IndicatorPosition(pos2)
	} else if pos3, ok3 := dto["position"].(IndicatorPosition); ok3 {
		position = pos3
	} else {
		panic("Unknown value type for position")
	}

	return &RandomStringStrategy{
		Position:   position,
		MinWords:   minWords,
		MaxWords:   maxWords,
		MinWordLen: 3,
		MaxWordLen: 6,
	}

}

func (f *RandomStringStrategy) Apply(content string, username string, id string) (string, []string) {
	indicators := f.GetIndicators()

	if f.Position == Beginning {
		return fmt.Sprintf("%s %s", strings.Join(indicators, " "), content), indicators
	} else if f.Position == Middle {
		wordCount := len(strings.Split(content, " "))
		halfCount := wordCount / 2
		return fmt.Sprintf("%s %s %s", content[:halfCount], strings.Join(indicators, " "), content[halfCount:]), indicators
	} else {
		return fmt.Sprintf("%s %s", content, strings.Join(indicators, " ")), indicators
	}
}

func (f *RandomStringStrategy) GetIndicators() []string {
	var indicators []string

	if f.MinWords <= 0 && f.MaxWords <= 0 {
		return []string{}
	}

	wordCount := rand.Intn(f.MaxWords-f.MinWords) + f.MinWords

	for i := 0; i < wordCount; i++ {
		strLen := rand.Intn(f.MaxWordLen-f.MinWordLen) + f.MinWordLen
		indicators = append(indicators, utils.GetRandomString(strLen))
	}

	return indicators
}
