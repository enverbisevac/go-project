package app

import (
	"os"
	"path"
)

const (
	AppDir  = "app"
	DataDir = "data"
)

func ConfigDir() (string, error) {
	config, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := path.Join(config, AppDir, DataDir)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return "", err
	}
	return dir, nil
}
