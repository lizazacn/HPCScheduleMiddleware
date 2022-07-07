package Service

import (
	"ScheduleSystem/Conf"
	"ScheduleSystem/Dao"
	"ScheduleSystem/ShellClient"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

var SqlConnInfo Dao.SqlConner

func Crontab() {
	SqlConnData := Dao.SQLConnInfo{
		User:        Conf.ServiceConfig.Database.Username,
		Password:    Conf.ServiceConfig.Database.Password,
		Host:        Conf.ServiceConfig.Database.Host,
		Port:        Conf.ServiceConfig.Database.Port,
		DbName:      Conf.ServiceConfig.Database.DBName,
		MaxIdleCons: Conf.ServiceConfig.Database.MaxIdleCons,
		MaxOpenCons: Conf.ServiceConfig.Database.MaxOpenCons,
		DBType:      Conf.ServiceConfig.Database.DBType,
		DBPath:      Conf.ServiceConfig.Database.Path,
	}
	var connInfo Dao.SqlConner
	if SqlConnData.DBType == 1 {
		connInfo = Dao.NewMySqlConner(SqlConnData)
	} else {
		connInfo = Dao.NewSqliteConner(SqlConnData)
	}
	//var connInfo Dao.SqlConner = Dao.NewMySqlConner(SqlConnData)
	SqlConnInfo = connInfo
	err := connInfo.CreateTable()
	if err != nil {
		log.Printf("创建数据表异常：%v", err)
		return
	}
	var lock sync.WaitGroup
	for _, val := range Conf.Config.ScheduleConfs {
		lock.Add(1)
		go getJob(val.AdminUser, val.AdminPassword, "", val.ScheduleName, "0", int(val.ScheduleID))
	}
	lock.Done()
}

// 获取作业信息
func getJob(username, password, path, ScheduleName, sessionId string, ScheduleID int) {
	var CrontabTime = Conf.ServiceConfig.Crontab.ExecInterval

	var unEndJobIDList []string
	var loginCount int
	ScheduleIdx := Conf.Config.ScheduleIDToIdx[ScheduleID]
	Schedule := Conf.Config.ScheduleConfs[ScheduleIdx]
	cmdStr := fmt.Sprintf("%s \"%s\"", Schedule.HistoryCommand, time.Now().Add(time.Duration(0-Schedule.HistoryTimeStep)*time.Minute).Format("2006-01-02T15:04:05"))
relogin:
	_, ConnID, err := Login(username, password, path, ScheduleName, sessionId, ScheduleID)
	if err != nil {
		loginCount++
		if loginCount >= 5 {
			log.Println("连续5次尝试建立连接失败，请检查连接信息和网络状态后重试！")
			return
		}
		time.Sleep(3 * time.Minute)
		goto relogin
	}
	loginCount = 0
callback:
	result, err := ShellClient.ShellConn.Send(ConnID, cmdStr)
	if err != nil {
		log.Println(err)
		//time.Sleep(time.Duration(CrontabTime) * time.Second)
		goto relogin
	}
	var offsetData = ""
	if len(unEndJobIDList) > 0 {
		offsetData, err = ShellClient.ShellConn.Send(ConnID, fmt.Sprintf("%s %s", Schedule.HistoryOffsetCommand, strings.Join(unEndJobIDList, ",")))
		if err != nil {
			log.Println(err)
			time.Sleep(time.Duration(CrontabTime) * time.Second)
			goto callback
		}
		unEndJobIDList = nil
	}

	if Schedule.HistoryResultType == 1 {
		var resultMap []map[string]interface{}
		err = json.Unmarshal([]byte(result), &resultMap)
		if err != nil {
			log.Println(err)
			time.Sleep(time.Duration(CrontabTime) * time.Second)
			goto callback
		}

		var ofsResultMap []map[string]interface{}
		err = json.Unmarshal([]byte(result), &ofsResultMap)
		if err != nil {
			log.Println(err)
			time.Sleep(time.Duration(CrontabTime) * time.Second)
			goto callback
		}
		resultMap = append(resultMap, ofsResultMap...)
		err, lastJobData := DataSort(&resultMap, &Schedule.HistoryToSql, &unEndJobIDList, strconv.Itoa(ScheduleID))
		if err != nil {
			time.Sleep(time.Duration(CrontabTime) * time.Second)
			return
		}

		lastJob := lastJobData.(map[string]interface{})
		cmdStr = fmt.Sprintf("%s \"%s\"", Schedule.HistoryCommand, lastJob["job_submit_time"])
		time.Sleep(time.Duration(CrontabTime) * time.Second)
		goto callback
	}
	procssResult, err := ProcssResult(result, Schedule.HistoryResultSeparator)
	if err != nil {
		log.Println(err)
		time.Sleep(time.Duration(CrontabTime) * time.Second)
		goto callback
	}
	// 补偿回调信息
	offsetResult, err := ProcssResult(offsetData, Schedule.HistoryResultSeparator)
	if err != nil {
		log.Println(err)
		time.Sleep(time.Duration(CrontabTime) * time.Second)
		goto callback
	}
	ofsResult := offsetResult.([]interface{})
	psResult := procssResult.([]interface{})
	psResult = append(psResult, ofsResult...)
	err, lastJobData := DataSort(&psResult, &Schedule.HistoryToSql, &unEndJobIDList, strconv.Itoa(ScheduleID))
	if err != nil {
		log.Println(err)
		time.Sleep(time.Duration(CrontabTime) * time.Second)
		goto callback
	}
	lastJob := lastJobData.(map[string]interface{})
	cmdStr = fmt.Sprintf("%s \"%s\"", Schedule.HistoryCommand, lastJob["job_submit_time"])
	time.Sleep(time.Duration(CrontabTime) * time.Second)
	goto callback
}

// DataSort 存储作业信息
func DataSort(resultMap interface{}, dataRelation *map[string]string, unEndJobIDList *[]string, scheduleId string) (error, interface{}) {
	var jobDataList []map[string]interface{}
	var lastJobData interface{}

	for _, m := range *(resultMap.(*[]interface{})) {
		var jobData = make(map[string]interface{})
		for k, val := range *dataRelation {
			jobData[k] = m.(map[string]string)[val]
		}
		jobData["schedule_id"] = scheduleId
		if jobData["job_use_gpus"] == "" {
			jobData["job_use_gpus"] = 0
		}
		if jobData["job_end_time"] == "Unknown" || jobData["job_end_time"] == "" {
			*unEndJobIDList = append(*unEndJobIDList, jobData["job_id"].(string))
		}
		jobDataList = append(jobDataList, jobData)
		lastJobData = jobData
	}
	if len(jobDataList) <= 0 {
		return errors.New("未解析到数据！"), nil
	}
	err := SqlConnInfo.InsertAndUpdateDataTable(jobDataList)
	if err != nil {
		log.Printf("插入数据异常：%v", err)
		return err, nil
	}
	return nil, lastJobData
}
