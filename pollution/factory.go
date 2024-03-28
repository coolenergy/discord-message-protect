package pollution

import "sync"

var instances = map[string]IPollutionStrategy{}
var instancesMutex = sync.Mutex{}

func GetPollutionStrategy(strategyName string, args map[string]interface{}) IPollutionStrategy {
	instancesMutex.Lock()
	if instance, found := instances[strategyName]; found {
		instancesMutex.Unlock()
		return instance
	} else {
		switch strategyName {
		case IncrementIntStrategyName:
			instance = newIncrementalStrategy(args)
			instances[strategyName] = instance
			instancesMutex.Unlock()
			return instance
		case FakerStrategyName:
			instance = newFakerStrategy(args)
			instances[strategyName] = instance
			instancesMutex.Unlock()
			return instance
		case RandomString:
			instance = newRandomStringStrategy(args)
			instances[strategyName] = instance
			instancesMutex.Unlock()
			return instance
		default:
			instancesMutex.Unlock()
		}
	}

	return nil
}
