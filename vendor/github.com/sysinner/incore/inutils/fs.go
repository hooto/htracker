package inutils

import (
	"errors"
	"os"
	"strings"
)

func FsMakeFileDir(path string, uid, gid int, mode os.FileMode) error {

	if idx := strings.LastIndex(path, "/"); idx > 0 {
		return FsMakeDir(path[0:idx], uid, gid, mode)
	}

	return nil
}

func FsMakeDir(path string, uid, gid int, mode os.FileMode) error {

	if _, err := os.Stat(path); err == nil {
		return nil
	}

	if uid < 500 || gid < 500 {
		return errors.New("Invalid uid or gid")
	}

	paths, path := strings.Split(strings.Trim(path, "/"), "/"), ""

	for _, v := range paths {

		path += "/" + v

		if _, err := os.Stat(path); err == nil {
			continue
		}

		if err := os.Mkdir(path, mode); err != nil {
			return err
		}

		os.Chmod(path, mode)
		os.Chown(path, uid, gid)
	}

	return nil
}
