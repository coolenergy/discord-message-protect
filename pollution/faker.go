package pollution

import (
	"fmt"
	"github.com/go-faker/faker/v4"
	"math/rand"
	"strings"
)

type FakerStrategy struct {
	Position IndicatorPosition
	MinWords int
	MaxWords int
}

func (f *FakerStrategy) GetName() string {
	return FakerStrategyName
}

type CreateFakerStrategyDto struct {
	Position   IndicatorPosition
	MinPadding int
	MaxPadding int
}

func newFakerStrategy(dto map[string]interface{}) *FakerStrategy {
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

	return &FakerStrategy{
		Position: position,
		MinWords: minWords,
		MaxWords: maxWords,
	}
}

func StringToPosition(position string) IndicatorPosition {
	if position == "trail" {
		return Trail
	} else if position == "beginning" {
		return Beginning
	} else if position == "random" {
		return Random
	} else if position == "middle" {
		return Middle
	}

	panic("Unknown argument " + position)
}

func (f *FakerStrategy) Apply(content string, username string, id string) (string, []string) {
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

func (f *FakerStrategy) GetIndicators() []string {
	var indicators []string

	if f.MinWords <= 0 && f.MaxWords <= 0 {
		return []string{}
	}

	min := f.MinWords
	max := f.MaxWords
	paddingCount := 0
	if min == max {
		paddingCount = min
	} else {
		paddingCount = rand.Intn(max-min) + min
	}

	for i := 0; i < paddingCount; i++ {
		indicators = append(indicators, faker.Word())
	}

	return indicators
}
