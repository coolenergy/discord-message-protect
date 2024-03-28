package pollution

import (
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
)

var counter uint64

type IncrementalStrategy struct {
	Position IndicatorPosition
}

func (i *IncrementalStrategy) GetName() string {
	return IncrementIntStrategyName
}

type CreateIncrementalStrategyDto struct {
	Position IndicatorPosition
}

func newIncrementalStrategy(dto map[string]interface{}) *IncrementalStrategy {
	position := StringToPosition(dto["position"].(string))
	return &IncrementalStrategy{
		Position: position,
	}
}

func (i *IncrementalStrategy) Apply(content string, username string, id string) (string, []string) {
	newId := atomic.AddUint64(&counter, 1)
	indicators := []string{
		strconv.Itoa(int(newId)),
	}

	if i.Position == Beginning {
		return fmt.Sprintf("%d %s", newId, content), indicators
	} else if i.Position == Middle {
		wordCount := len(strings.Split(content, " "))
		halfCount := wordCount / 2
		return fmt.Sprintf("%s %d %s", content[:halfCount], newId, content[halfCount:]), indicators
	} else {
		return fmt.Sprintf("%s %d", content, newId), indicators
	}
}
