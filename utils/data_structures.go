package utils

// RotateMap is used to rotate a map, although it seems useless, in Go deleted pairs in a map
// are not deleted in memory, the map keeps its size, as the app adds/removes pairs
// the map keeps growing, wasting memory and potentially crashing due to Out of memory errors
// this method takes a map that we have used heavily, and makes a copy of it without those hidden yet present
// pairs that we no longer use.
func RotateMap[K comparable, V any](someMap map[K]V) map[K]V {
	newMap := make(map[K]V, len(someMap))
	for a, b := range someMap {
		newMap[a] = b
	}

	return newMap
}
