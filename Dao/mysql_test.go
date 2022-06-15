package Dao

import "testing"

func TestConnInfo_CreateTable(t *testing.T) {
	connInfo := &ConnInfo{
		User:        "root",
		Password:    "123456",
		Host:        "127.0.0.1",
		Port:        3306,
		DbName:      "schedule",
		MaxIdleCons: 5,
		MaxOpenCons: 5,
	}
	connInfo.Init()
	connInfo.CreateTable()
}
