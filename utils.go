package fsarch

import (
	"errors"
	"os"
	"strconv"
	"time"
)

func Touch(fileName string) error {
	if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
		file, err := os.Create(fileName)
		if err != nil {
			return err
		}
		_ = file.Close()
		return nil
	}
	now := time.Now().Local()
	return os.Chtimes(fileName, now, now)
}

func CheckName(fileName string) bool {
	if len(fileName) < 2 {
		return false
	}
	_, err := strconv.ParseInt(fileName[:2], 16, 32)
	return err == nil
}
