package utils

import "os"

func DirExists(dirPath string) bool {

	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) || info == nil {
		return false
	}
	return info.IsDir()
}
