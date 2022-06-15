package Service

import (
	"ScheduleSystem/Conf"
	"ScheduleSystem/ShellClient"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

// Init 初始化读取配置文件
func Init(path string) {
	serviceConf, err := Conf.ReadServiceConf(path)
	if err != nil {
		return
	}
	Conf.ServiceConfig = serviceConf
	c, err := Conf.ReadConf(serviceConf.Schedule.ConfPath)
	if err != nil {
		log.Printf("读取配置文件异常！")
		return
	}
	Conf.Config = c
}

// Login 登录SN节点 参数为：用户名、密码、工作路径、调度系统名称、会话ID、调度系统ID；
// 提示：ScheduleName和ScheduleID存在一个即可，同时存在时ScheduleID优先级大于ScheduleName
// 响应：调度系统ID、连接ID、 error
func Login(username, password, path, ScheduleName, sessionId string, ScheduleID int) (uint, string, error) {
	if ScheduleName == "" && ScheduleID == 0 {
		log.Printf("传递参数异常，调度系统名称和调度系统ID不能同时为空！")
		return 0, "", nil
	}
	if ScheduleName != "" && ScheduleID == 0 {
		scheIds := Conf.Config.ScheduleNameToId[ScheduleName]
		idx := rand.Intn(len(scheIds))
		ScheduleID = scheIds[idx]
	}
	i := Conf.Config.ScheduleIDToIdx[ScheduleID]
	scheduleConf := Conf.Config.ScheduleConfs[i]
	host := scheduleConf.LoginHosts[rand.Intn(len(scheduleConf.LoginHosts))]
	id := md5.Sum([]byte(username + password + host + strconv.Itoa(ScheduleID) + sessionId))
	strId := hex.EncodeToString(id[:])
	args := strings.Split(host, ":")
	if len(args) != 2 {
		return scheduleConf.ScheduleID, strId, errors.New("登录节点地址格式异常！")
	}
	ip := args[0]
	port, err := strconv.Atoi(args[1])
	if err != nil {
		return scheduleConf.ScheduleID, strId, errors.New("登录节点地址格式异常！")
	}
	if ShellClient.ShellConn == nil {
		ShellClient.ShellConn = ShellClient.Init()
	}
	code, err := ShellClient.ShellConn.CreateSessionConn(strId, ip, username, password, port)
	if err != nil {
		return scheduleConf.ScheduleID, strId, err
	}
	switch code {
	case FAIL:
		fmt.Println("Fail!")
		break
	case SUCCESS:
		fmt.Println("Success!")
		break
	case MAXIMUM_CONNECTIONS_EXECEEDED:
		fmt.Println("Maximum connections exceeded")
		break
	case CONNECTION_ALREADY_EXISTS:
		fmt.Println("Connection already exists")
		break
	default:
		fmt.Println("Fail!")
	}
	_, err = ShellClient.ShellConn.Send(strId, "cd "+path)
	if err != nil {
		return scheduleConf.ScheduleID, strId, err
	}
	return scheduleConf.ScheduleID, strId, nil
}

// Exec 执行命令， 参数：连接ID、命令、调度系统ID、命令ID（不知道可写0）、命令参数
// 提示：cmd和CommandID存在一个即可，同时存在时CommandID优先级大于cmd
// 响应：error、执行命令后的返回值
func Exec(id, cmd string, ScheduleID int, CommandID uint, args ...string) (error, interface{}) {
	if ScheduleID <= 0 {
		log.Printf("传递参数异常！")
		return errors.New("调度ID参数异常！"), nil
	}

	ScheduleIdx := Conf.Config.ScheduleIDToIdx[ScheduleID]
	Schedule := Conf.Config.ScheduleConfs[ScheduleIdx]
	if CommandID <= 0 {
		CommandID = Schedule.CommandToID[cmd]
		if CommandID <= 0 {
			log.Printf("没有当前命令！")
			return errors.New("命令ID参数异常！"), nil
		}
	}
	cmdArgs := strings.Join(args, " ")
	idx := Schedule.CommandIDToIdx[CommandID]
	command := Schedule.Cmd[idx]
	result, err := ShellClient.ShellConn.Send(id, fmt.Sprintf("%s %s %s", command.Cmd, command.Args, cmdArgs))
	if err != nil {
		return errors.New("执行命令异常！"), nil
	}

	if command.ResultType == 0 && command.ResultUseJson == 1 {
		procssResult, err := ProcssResult(result, command.ResultSeparator)
		if err != nil {
			return err, nil
		}
		return nil, procssResult
	}

	return nil, result
}

// ProcssResult 处理响应结果， 参数：返回字符串，响应结果分隔符
func ProcssResult(str, Separator string) (interface{}, error) {
	var err error
	lines := strings.Split(str, "\n")
	var results []interface{}
	if Separator != "" {
		title := strings.Split(lines[0], Separator)
		for _, line := range lines[1:] {
			if len(line) < len(title) {
				continue
			}
			resultMap := make(map[string]string)
			titleList := strings.Split(line, Separator)
			for idx, t := range title {
				resultMap[t] = titleList[idx]
			}
			results = append(results, resultMap)
		}
	} else {
		re, err := regexp.Compile(`\w*\s*|\w*\d*\w*\s*|\d*\w*\d*\s*`)
		if err != nil {
			return nil, err
		}
		title := re.FindAllString(lines[0], -1)
		index := re.FindAllIndex([]byte(lines[0]), -1)
		for _, line := range lines[1:] {
			if len(line) < index[len(index)-1][0] {
				continue
			}
			resultMap := make(map[string]string)
			//titleList := re.FindAllString(line, -1)
			for idx, t := range title {
				if idx < len(title)-1 {
					resultMap[t] = line[index[idx][0]:index[idx+1][0]]
					continue
				}
				resultMap[t] = line[index[idx][0]:]
			}
			results = append(results, resultMap)
		}
	}
	return results, err
}
