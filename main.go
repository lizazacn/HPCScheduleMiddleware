package main

import (
	"ScheduleSystem/API"
	"ScheduleSystem/Conf"
	"ScheduleSystem/Service"
	"ScheduleSystem/rpc"
	"fmt"
	"log"
)

func main() {
	Service.Init("./Conf/service.json")
	routes := API.Routes()
	// 定时任务
	if Conf.ServiceConfig.Crontab.Status {
		go Service.Crontab()
	}
	// RPC服务端
	go rpc.Run(fmt.Sprintf("%s:%d", Conf.ServiceConfig.Rpc.Host, Conf.ServiceConfig.Rpc.Port))
	// Http服务端
	err := routes.Run(fmt.Sprintf("%s:%d", Conf.ServiceConfig.Http.Host, Conf.ServiceConfig.Http.Port))
	if err != nil {
		log.Printf("Http服务监听异常：%v", err)
		return
	}
}
