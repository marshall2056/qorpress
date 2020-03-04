package helpers

import (
	"os"
	"path/filepath"
)

func FileExistsFn(basepath, filename string) bool {
	if _, err := os.Stat(filepath.Join(basepath, filename)); os.IsNotExist(err) {
		return false
	}
	return true
}
