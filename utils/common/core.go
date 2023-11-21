package common

import "path/filepath"

func GetFileName(uri string) string {
	_, fileName := filepath.Split(uri)
	return fileName
}
