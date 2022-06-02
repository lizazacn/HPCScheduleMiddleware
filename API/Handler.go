package API

import (
	"ScheduleSystem/Service"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type InitInfo struct {
	UserName		string `json:"username"`
	Password		string `json:"password"`
	Path			string `json:"path"`
	ScheduleName 	string `json:"schedule_name"`
	SessionID		string `json:"session_id"`
	ScheduleID		int `json:"schedule_id"`
}

type ExecInfo struct {
	ConnID 		string 	`json:"conn_id"`
	Cmd 		string 	`json:"cmd"`
	ScheduleID 	int 	`json:"schedule_id"`
	CommandID 	uint 	`json:"command_id"`
	Args 		string 	`json:"args"`
}

func Init(ctx *gin.Context){
	initInfo := new(InitInfo)
	err := ctx.Bind(initInfo)
	if err != nil {
		log.Printf("解析请求信息异常：%v", err)
		ctx.JSON(http.StatusOK, gin.H{
			"code": StatusServerErr,
			"msg": "请求参数异常！",
			"data": "",
		})
		return
	}
	scheduleID, connID, err := Service.Login(initInfo.UserName, initInfo.Password, initInfo.Path, initInfo.ScheduleName, initInfo.SessionID, initInfo.ScheduleID)
	if err != nil {
		log.Printf("初始化客户端连接异常！")
		ctx.JSON(http.StatusOK, gin.H{
			"code": StatusServerErr,
			"msg": "初始化客户端连接异常！",
			"data": "",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code": StatusOK,
		"msg": "初始化连接成功！",
		"data": gin.H{
			"ScheduleID": scheduleID,
			"ConnID": connID,
		},
	})
}

func Exec(ctx *gin.Context){
	execInfo := new(ExecInfo)
	err := ctx.Bind(execInfo)
	if err != nil {
		log.Printf("解析请求参数异常：%v", err)
		ctx.JSON(http.StatusOK, gin.H{
			"code": StatusServerErr,
			"msg": "请求参数异常！",
			"data": "",
		})
		return
	}

	err, result := Service.Exec(execInfo.ConnID, execInfo.Cmd, execInfo.ScheduleID, execInfo.CommandID, execInfo.Args)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code": StatusServerErr,
			"msg": "执行命令异常！",
			"data": "",
		})
		log.Printf("命令执行异常：%v", err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code": StatusOK,
		"msg":"命令执行成功！",
		"data": result,
	})
}
