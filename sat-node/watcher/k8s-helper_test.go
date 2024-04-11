package watcher

import (
	"github.com/stretchr/testify/assert"
	"satweave/utils/logger"
	"testing"
)

func TestGetAllNamespaces(t *testing.T) {
	namespaces, err := GetAllNamespaces()
	if err != nil {
		logger.Errorf("Failed to get namespaces: %v", err)
	}
	assert.NoError(t, err)
	if len(namespaces) == 0 {
		logger.Errorf("Expected to get at least one namespace, got 0")
	}
	assert.NotEqual(t, 0, len(namespaces))
	// 打印获取到的命名空间，仅用于调试
	for _, ns := range namespaces {
		logger.Infof("Namespace: %s", ns)
	}
}
