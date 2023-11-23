package utils

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

func MkDirIfNotExist(dirPath string) {
	err := os.Mkdir(dirPath, 0755)

	if err != nil && !errors.Is(err, fs.ErrExist) {
		fmt.Println(err)
		os.Exit(1)
	}
}
