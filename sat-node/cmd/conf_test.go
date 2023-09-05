package cmd

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"satweave/sat-node/config"
	config2 "satweave/utils/config"
	"testing"
)

func TestRealConfig(t *testing.T) {
	conf := config.Config{}
	err := config2.TransferJsonToConfig(&conf, "sat_node.json")
	assert.NoError(t, err)

	fmt.Println(conf)
}
