package Dao

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type MySqlConnInfo struct {
	DriverName  string   //驱动名称
	User        string   // 连接用户名
	Password    string   // 连接密码
	Host        string   // mysql服务器地址
	Port        int      // mysql服务端口号
	DbName      string   // 数据库名
	DBConn      *sqlx.DB // 数据库连接
	MaxOpenCons int      // 最大连接数
	MaxIdleCons int      // 最大闲置连接
}

type JobInfo struct {
	Cluster         string `json:"cluster" db:"cluster"`                     // 集群名称
	ScheduleID      string `json:"schedule_id" db:"schedule_id"`             // 调度系统ID
	JobID           string `json:"job_id" db:"job_id"`                       // 作业ID
	JobName         string `json:"job_name" db:"job_name"`                   // 作业名称
	JobAccount      string `json:"job_account" db:"job_account"`             // 作业运行用户
	JobAccountGroup string `json:"job_account_group" db:"job_account_group"` // 作业运行用户组
	JobQueue        string `json:"job_queue" db:"job_queue"`                 // 作业队列
	JobStatus       string `json:"job_status" db:"job_status"`               // 作业状态
	JobUseNodes     string `json:"job_use_nodes" db:"job_use_nodes"`         // 作业使用节点数
	JobNodeList     string `json:"job_node_list" db:"job_node_list"`         // 作业使用节点列表
	JobUseCPUS      string `json:"job_use_cpus" db:"job_use_cpus"`           // 作业使用的核心数
	JobUseGPUS      string `json:"job_use_gpus" db:"job_use_gpus"`           // 作业使用的GPU数
	JobExecCommand  string `json:"job_exec_command" db:"job_exec_command"`   // 作业运行脚本
	JobSubmitTime   string `json:"job_submit_time" db:"job_submit_time"`     // 作业提交时间
	JobStartTime    string `json:"job_start_time" db:"job_start_time"`       // 作业开始时间
	JobEndTime      string `json:"job_end_time" db:"job_end_time"`           // 作业结束运行时间
	JobRunningTime  string `json:"job_running_time" db:"job_running_time"`   // 作业运行时长
	JobWorkDir      string `json:"job_work_dir" db:"job_work_dir"`           // 作业工作目录
}

func NewMySqlConner(connInfo SQLConnInfo) *MySqlConnInfo {
	connInfo.DriverName = "mysql"
	db, err := sqlx.Open(connInfo.DriverName, fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", connInfo.User, connInfo.Password, connInfo.Host, connInfo.Port, connInfo.DbName))
	if err != nil {
		log.Println("打开时数据库连接失败")
		return nil
	}
	db.SetMaxOpenConns(connInfo.MaxOpenCons)
	db.SetMaxIdleConns(connInfo.MaxIdleCons)

	err = db.Ping()
	if err != nil {
		return nil
	}
	connInfo.DBConn = db
	return &MySqlConnInfo{
		DriverName:  connInfo.DriverName,
		DBConn:      connInfo.DBConn,
		User:        connInfo.User,
		Password:    connInfo.Password,
		Host:        connInfo.Host,
		Port:        connInfo.Port,
		DbName:      connInfo.DbName,
		MaxIdleCons: connInfo.MaxIdleCons,
		MaxOpenCons: connInfo.MaxOpenCons,
	}
}

func (connInfo *MySqlConnInfo) CreateTable() error {
	sql := "CREATE TABLE IF NOT EXISTS `schedule_jobs` (\n\t`id` INT NOT NULL AUTO_INCREMENT COMMENT 'ID'," +
		"\n\t`cluster` VARCHAR(60) NULL DEFAULT NULL COMMENT '集群名称'," +
		"\n\t`schedule_id` VARCHAR(60) NULL DEFAULT NULL COMMENT '调度系统ID'," +
		"\n\t`job_id` VARCHAR(60) NULL DEFAULT NULL COMMENT '作业ID'," +
		"\n\t`job_name` VARCHAR(60) NULL DEFAULT NULL COMMENT '作业名'," +
		"\n\t`job_account` VARCHAR(60) NULL DEFAULT NULL COMMENT '作业提交用户'," +
		"\n\t`job_account_group` VARCHAR(60) NULL DEFAULT NULL COMMENT '作业提交用户组'," +
		"\n\t`job_queue` VARCHAR(60) NULL DEFAULT NULL COMMENT '作业队列'," +
		"\n\t`job_status` VARCHAR(60) NULL DEFAULT NULL COMMENT '作业状态'," +
		"\n\t`job_use_nodes` BIGINT NULL DEFAULT '0' COMMENT '作业占用节点数'," +
		"\n\t`job_node_list` TEXT NULL DEFAULT NULL COMMENT '作业占用节点列表'," +
		"\n\t`job_use_cpus` BIGINT NULL DEFAULT '0' COMMENT '作业占用核心数'," +
		"\n\t`job_use_gpus` BIGINT NULL DEFAULT '0' COMMENT '作业占用gpu数'," +
		"\n\t`job_exec_command` TEXT NULL DEFAULT NULL COMMENT '作业执行脚本'," +
		"\n\t`job_submit_time` VARCHAR(60) NULL DEFAULT NULL COMMENT '作业提交时间'," +
		"\n\t`job_start_time` VARCHAR(60) NULL DEFAULT NULL COMMENT '作业开始运行时间'," +
		"\n\t`job_end_time` VARCHAR(60) NULL DEFAULT NULL COMMENT '作业结束运行时间'," +
		"\n\t`job_running_time` VARCHAR(60) NULL DEFAULT NULL COMMENT '作业运行时长'," +
		"\n\t`job_work_dir` TEXT NULL DEFAULT NULL COMMENT '作业工作目录'," +
		"\n\t`create_time` DATETIME NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间'," +
		"\n\t`deleted_status` INT(11) NULL DEFAULT '0' COMMENT '记录删除状态'," +
		"\n\tUNIQUE INDEX `UNIQUEKEY` (`job_id`, `schedule_id`) USING BTREE," +
		"\n\tINDEX `key` (`id`)\n)\nCOLLATE='utf8_bin'\n;"
	_, err := connInfo.DBConn.Exec(sql)
	if err != nil {
		return err
	}
	return nil
}

func (connInfo *MySqlConnInfo) InsertAndUpdateDataTable(Data interface{}) error {
	sql := "INSERT INTO schedule_jobs(cluster,schedule_id," +
		"job_id,job_name,job_account,job_account_group,job_queue," +
		"job_status,job_use_nodes,job_node_list,job_use_cpus,job_use_gpus," +
		"job_exec_command," +
		"job_submit_time,job_start_time,job_end_time,job_running_time," +
		"job_work_dir) " +
		"VALUES(:cluster,:schedule_id," +
		":job_id,:job_name,:job_account,:job_account_group,:job_queue," +
		":job_status,:job_use_nodes,:job_node_list,:job_use_cpus,:job_use_gpus," +
		":job_exec_command," +
		":job_submit_time,:job_start_time,:job_end_time,:job_running_time," +
		":job_work_dir) " +
		"ON DUPLICATE KEY UPDATE " + // 后边为更新数据的内容
		"job_start_time=VALUES(job_start_time)," +
		"job_end_time=VALUES(job_end_time)," +
		"job_status=VALUES(job_status);"

	_, err := connInfo.DBConn.NamedExec(sql, Data.([]map[string]interface{}))
	if err != nil {
		return err
	}
	return nil
}
