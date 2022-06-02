package Conf

import "testing"

func TestReadServiceConf(t *testing.T) {
	conf, err := ReadServiceConf("service.json")
	if err != nil {
		return
	}
	t.Log(conf)
}
