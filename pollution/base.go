package pollution

type IPollutionStrategy interface {
	Apply(content string, username string, id string) (string, []string)
	GetName() string
}

type IndicatorPosition string

const (
	Beginning IndicatorPosition = "beginning"
	Trail     IndicatorPosition = "trail"
	Middle    IndicatorPosition = "middle"
	Random    IndicatorPosition = "random"

	IncrementIntStrategyName string = "incremental_inc"
	FakerStrategyName        string = "faker"
	RandomString             string = "random_string"
)
