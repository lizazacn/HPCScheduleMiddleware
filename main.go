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
	go rpc.Run(fmt.Sprintf("%s:%d", Conf.ServiceConfig.Rpc.Host, Conf.ServiceConfig.Rpc.Port))
	err := routes.Run(fmt.Sprintf("%s:%d", Conf.ServiceConfig.Http.Host, Conf.ServiceConfig.Http.Port))
	if err != nil {
		log.Printf("Http服务监听异常：%v", err)
		return
	}
}
