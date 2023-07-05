package config

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestGetConf(t *testing.T) {
	type TestSubConf struct {
		Name string
		Age  uint64
	}
	type TestConf struct {
		Config
		Name  string
		Age   uint64
		Child TestSubConf
	}
	conf := &TestConf{
		Name: "SincereXIA",
		Age:  22,
		Child: TestSubConf{
			Name: "XiongHC",
			Age:  22,
		},
	}

	var testConf TestConf
	Register(&testConf, "test.json")
	ReadAll()

	if !reflect.DeepEqual(&testConf, conf) {
		t.Errorf("config not equal")
	}
}

func TestTransferJsonToConfig(t *testing.T) {
	type TestSubConf struct {
		Name string `json:"Name"`
		Age  uint64 `json:"Age"`
	}
	type TestConf struct {
		Name  string      `json:"Name"`
		Age   uint64      `json:"Age"`
		Child TestSubConf `json:"Child"`
	}
	conf := &TestConf{
		Name: "SincereXIA",
		Age:  22,
		Child: TestSubConf{
			Name: "XiongHC",
			Age:  22,
		},
	}

	var testConf TestConf
	err := TransferJsonToConfig(&testConf, "test.json")
	assert.NoError(t, err)

	fmt.Println(testConf)
	fmt.Println(conf)

	if !reflect.DeepEqual(&testConf, conf) {
		t.Errorf("config not equal")
	}

}
