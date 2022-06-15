package Conf

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Conf json配置文件格式
type Conf struct {
	ScheduleConfs    []ScheduleConf   `json:"sche_sys_conf"` // 调度系统配置
	ShellInfo        ShellConnInfo    `json:"shell_info"`    // shell连接相关信息
	ScheduleIDToIdx  map[int]int      // 调度ID与json index的对应关系
	ScheduleNameToId map[string][]int // 调度名与json index的对应关系
}

type ServiceConf struct {
	Http     HTTP     `json:"http"`     // http协议相关配置
	Rpc      RPC      `json:"rpc"`      // rpc协议相关配置
	Schedule SCHEDULE `json:"schedule"` // 调度相关配置（目前只有配置调度配置文件的存放路径）
	Database DATABASE `json:"database"` // 数据库相关配置（用于存放调度系统作业信息）
	Crontab  CRONTAB  `json:"crontab"`  // 定时任务配置
}

//// ScheduleIDToIdx 调度ID与json index的对应关系
//type ScheduleIDToIdx map[int]int
//
//// ScheduleNameToIdx 调度名与json index的对应关系
//type ScheduleNameToIdx map[string][]int

type ShellConnInfo struct {
	Timeout    int64 `json:"timeout"`     // 长连接最大保持时间 单位为：s
	MaxConn    int64 `json:"max_conn"`    // 最大连接数
	ListenStep int64 `json:"listen_step"` // 监听步长 单位为：ms
}

// ScheduleConf 单个调度系统对应的配置结构体
type ScheduleConf struct {
	LoginHosts             []string          `json:"login_hosts"`              // 登录节点地址数组，每项数据格式为127.0.0.1:22
	ScheduleName           string            `json:"schedule_name"`            // 调度系统名称
	ScheduleID             uint              `json:"schedule_id"`              // 调度系统ID
	Cmd                    []Command         `json:"command"`                  // 命令结构体数组
	AdminUser              string            `json:"admin_user"`               // 监听定时任务管理员账户
	AdminPassword          string            `json:"admin_password"`           // 监听定时任务管理员密码
	HistoryResultType      uint              `json:"history_result_type"`      // 结果数据类型（0：表格结构，1：表示json结构）默认值为0
	HistoryResultSeparator string            `json:"history_result_separator"` // 历史作业结果分割符（根据某种格式分割结果数据）
	HistoryTimeStep        int64             `json:"history_time_step"`        // 拉取历史作业时间步长
	HistoryCommand         string            `json:"history_command"`          // 查看历史作业命令
	HistoryOffsetCommand   string            `json:"history_offset_command"`   // 通过作业ID查看历史作业参数名
	HistoryToSql           map[string]string `json:"history_to_sql"`           // 历史作业相关字段与数据库字段绑定关系
	CommandIDToIdx         map[uint]int      // 命令类型与命令结构体数组关系
	CommandToID            map[string]uint   // 命令名称与类型的对应关系
}

// Command 调度命令结构体
type Command struct {
	Cmd             string `json:"cmd"`              // 作业相关命令
	Args            string `json:"args"`             // 强制性参数
	ResultSeparator string `json:"result_separator"` // 结果分割符（根据某种格式分割结果数据）
	CommandID       uint   `json:"command_id"`       // 调度系统命令类型（目前支持：1、交互式提交作业，2、后台提交作业，
	//3、获取队列信息，4、获取当前用户正在运行的作业， 5、获取历史作业， 【1-5表示固定命令类型，6-10表示其它命令类型】）
	ResultType    uint `json:"result_type"`     // 结果数据类型（0：表格结构，1：表示json结构，2：其它类型）默认值为2
	ResultUseJson uint `json:"result_use_json"` // 结果是否使用json格式化输出（0：表示原样输出，1：表示格式化输出），当result_type为0时有效
}

// HTTP http监听相关配置
type HTTP struct {
	Host string `json:"host"` // 监听地址
	Port int    `json:"port"` // 监听端口号
}

// RPC rpc监听相关配置
type RPC struct {
	Host string `json:"host"` // 监听地址
	Port int    `json:"port"` // 监听端口号
}

// SCHEDULE 调度相关配置
type SCHEDULE struct {
	ConfPath string `json:"conf_path"` // 调度系统配置文件路径
}

// DATABASE 数据库相关配置
type DATABASE struct {
	Status      bool   `json:"status"`        // 是否存储作业数据（此项参数与crontab中的status同时为true时生效）
	DBType      int    `json:"db_type"`       // 数据库类型，1、表示Mysql；2、表示sqlite
	Path        string `json:"path"`          // db文件存放路径
	Host        string `json:"host"`          // 数据库地址
	Port        int    `json:"port"`          // 数据库端口号
	DBName      string `json:"db_name"`       // 数据库名称
	Username    string `json:"username"`      // 用户名
	Password    string `json:"password"`      // 密码
	MaxOpenCons int    `json:"max_open_cons"` // 最大连接数
	MaxIdleCons int    `json:"max_idle_cons"` // 最大闲置连接
}

// CRONTAB 定时任务相关配置
type CRONTAB struct {
	Status       bool  `json:"status"`        // 定时任务状态
	ExecInterval int64 `json:"exec_interval"` // 定时任务执行间隔
}

var (
	Config        *Conf
	ServiceConfig *ServiceConf
)

func ReadServiceConf(args ...string) (*ServiceConf, error) {
	var path string
	if len(args) == 0 || args[0] == "" {
		path = "etc/service.json"
	} else {
		path = args[0]
	}

	confBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	serviceConf := new(ServiceConf)

	err = json.Unmarshal(confBytes, serviceConf)
	if err != nil {
		return nil, err
	}
	return serviceConf, nil
}

// ReadConf 从json中读取调度相关配置
func ReadConf(args ...string) (*Conf, error) {
	var path string
	if len(args) == 0 || args[0] == "" {
		path = "etc/config.json"
	} else {
		path = args[0]
	}
	confFile, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	confBytes, err := ioutil.ReadAll(confFile)
	if err != nil {
		return nil, err
	}
	conf := new(Conf)
	err = json.Unmarshal(confBytes, conf)
	if err != nil {
		return nil, err
	}
	for idx, val := range conf.ScheduleConfs {
		if val.LoginHosts == nil || len(val.LoginHosts) == 0 {
			continue
		}
		if conf.ScheduleIDToIdx == nil {
			conf.ScheduleIDToIdx = make(map[int]int)
		}
		conf.ScheduleIDToIdx[int(val.ScheduleID)] = idx
		if conf.ScheduleNameToId == nil {
			conf.ScheduleNameToId = make(map[string][]int)
		}
		conf.ScheduleNameToId[val.ScheduleName] = append(conf.ScheduleNameToId[val.ScheduleName], int(val.ScheduleID))

		for idx2, val2 := range val.Cmd {
			if val.CommandIDToIdx == nil {
				val.CommandIDToIdx = make(map[uint]int)
			}
			val.CommandIDToIdx[val2.CommandID] = idx2
		}
	}
	return conf, nil
}
