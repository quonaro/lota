package config

import (
	"fmt"
	"lota/shared"
	"os"
	"path/filepath"
)

type FileConfig struct {
	Path string
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func CurrentDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current dir: %w", err)
	}
	return dir, nil
}

func GetConfigPath(path string) (*FileConfig, error) {
	if path == "" {
		dir, err := CurrentDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(dir, shared.ConfigFileName)
	} else if isDir(path) {
		path = filepath.Join(path, shared.ConfigFileName)
	}
	return &FileConfig{Path: path}, nil
}
