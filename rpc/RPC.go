package rpc

import (
	"ScheduleSystem/Service"
	"fmt"
	"log"
	"net"
	"net/rpc"
)

type Schedule struct{
}

type InitRequest struct {
	UserName		string `json:"username"`
	Password		string `json:"password"`
	Path			string `json:"path"`
	ScheduleName 	string `json:"schedule_name"`
	SessionID		string `json:"session_id"`
	ScheduleID		int `json:"schedule_id"`
}

type InitResponse struct {
	Code int
	Message string
	Data interface{}
}

type ExecRequest struct {
	ConnID 		string 	`json:"conn_id"`
	Cmd 		string 	`json:"cmd"`
	ScheduleID 	int 	`json:"schedule_id"`
	CommandID 	uint 	`json:"command_id"`
	Args 		string 	`json:"args"`
}

type ExecResponse struct {
	Code int
	Message string
	Data interface{}
}

func (s Schedule)Init(req InitRequest, resp *InitResponse) error{
	scheduleID, connID, err := Service.Login(req.UserName, req.Password, req.Path, req.ScheduleName, req.SessionID, req.ScheduleID)
	if err != nil {
		log.Printf("初始化客户端连接异常:%v", err)
		resp.Code = StatusServerErr
		resp.Message = "初始化连接异常！"
		resp.Data = nil
		return err
	}
	resp.Code = StatusOK
	resp.Message = "连接初始化成功！"
	var data = make(map[string]interface{})
	data["ScheduleID"] = scheduleID
	data["ConnID"] = connID
	resp.Data = data
	return nil
}

func (s Schedule)Exec(req ExecRequest, resp *ExecResponse) error{
	err, result := Service.Exec(req.ConnID, req.Cmd, req.ScheduleID, req.CommandID, req.Args)
	if err != nil {
		log.Printf("执行命令异常:%v", err)
		resp.Code = StatusServerErr
		resp.Message = "执行命令异常！"
		resp.Data = nil
		return err
	}
	resp.Code = StatusOK
	resp.Message = "执行命令异常！"
	resp.Data = result
	return nil
}

func Run(addr string) error{
	err := rpc.Register(new(Schedule))
	if err != nil {
		log.Printf("启动RPC监听异常：%v", err)
		return err
	}
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("创建TCP监听异常：%v", err)
		return err
	}
	fmt.Println("Listening and serving RPC on "+addr)
	rpc.Accept(listen)
	return nil
}
