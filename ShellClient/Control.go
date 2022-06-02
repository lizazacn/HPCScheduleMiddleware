package ShellClient

import (
	"ScheduleSystem/Conf"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

// ShellConnect shell连接结构体
type ShellConnect struct {
	Timeout int64
	MaxConn int64
	NowConn int64
	ConnMap sync.Map
	sync.RWMutex
}

var ShellConn *ShellConnect

func Init() *ShellConnect {
	if Conf.Config == nil{
		conf, err := Conf.ReadConf()
		if err != nil{
			log.Fatalf("读取配置文件异常：%v", err)
			return nil
		}
		Conf.Config = conf
	}
	shellInfo := Conf.Config.ShellInfo
	con := &ShellConnect{
		Timeout: shellInfo.Timeout,
		MaxConn: shellInfo.MaxConn,
		NowConn: 0,
	}

	go con.Listener()

	return con
}

func (con *ShellConnect)CreateSessionConn(id, host, user, password string, port int) (int, error){
	var client *Cli
	if con.NowConn >= con.MaxConn{
		return MAXIMUM_CONNECTIONS_EXECEEDED, errors.New("超出最大连接数！")
	}
	if val, ok := con.ConnMap.Load(id); ok && val != nil {
		if val.(*Cli).SshSession == nil{
			client = val.(*Cli)
			err := client.Session()
			if err != nil {
				return FAIL, err
			}

			go client.Std()
		}
		return CONNECTION_ALREADY_EXISTS, errors.New("连接已存在！")
	}
	client = NewClient(host, port, user, password,nil)

	// 建立连接
	err := client.Connect()
	if err != nil {
		return FAIL, err
	}

	con.Lock()
	con.NowConn ++
	con.Unlock()
	con.ConnMap.Store(id, client)

	// 生成session
	err = client.Session()
	if err != nil {
		return FAIL, err
	}

	go client.Std()

	return SUCCESS, nil
}

func (con *ShellConnect)Send(id, cmd string) (out string, err error){
	if val, ok := con.ConnMap.Load(id); ok{
		if cmd == "closeTerminal"{
			val.(*Cli).UploadTime = time.Now().Add(time.Duration(con.Timeout << 1) * time.Second)
			return
		}
		outStr, errStr := val.(*Cli).Send(cmd)
		out = outStr + errStr
		return
	}
	err = errors.New("当前用户未建立连接！")
	return
}

// CreateCombinedOutputConn 创建一个连接，发送一次命令并获取返回值信息
func (con *ShellConnect)CreateCombinedOutputConn(id, host, user, password, cmd string, port int)(int, []byte, error){
	if con.NowConn >= con.MaxConn{
		return MAXIMUM_CONNECTIONS_EXECEEDED, nil, errors.New("超出最大连接数！")
	}
	if _, ok := con.ConnMap.Load(id); ok{
		return CONNECTION_ALREADY_EXISTS, nil, errors.New("连接已存在！")
	}
	client := NewClient(host, port, user, password,nil)

	// 建立连接
	err := client.Connect()
	if err != nil {
		return FAIL,nil,err
	}

	con.Lock()
	con.NowConn ++
	con.Unlock()
	con.ConnMap.Store(id, client)

	// 生成session
	err = client.Session()
	if err != nil {
		return FAIL, nil, err
	}

	output, err := client.SshSession.CombinedOutput(cmd)
	if err != nil {
		return FAIL, nil, err
	}
	client.SshSession.Close()
	client.SshSession = nil
	return SUCCESS, output, nil
}

// Listener 连接监听
func (con *ShellConnect) Listener() {
	fmt.Println("开始超时监听")
	for true {
		time.Sleep(time.Duration(Conf.Config.ShellInfo.ListenStep) * time.Millisecond)
		con.ConnMap.Range(func(key, value interface{}) bool {
			if value == nil {
				return false
			}
			if time.Now().Sub(value.(*Cli).UploadTime).Seconds() >= float64(con.Timeout){
				con.ConnMap.Delete(key)
				value.(*Cli).Send("closeTerminal")
				con.Lock()
				con.NowConn --
				con.Unlock()
			}
			return true
		})
	}
}
