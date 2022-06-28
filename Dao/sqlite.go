package Dao

import (
	"github.com/jmoiron/sqlx"
	"log"
)

type SqliteInfo struct {
	DriverName  string   //驱动名称
	DBConn      *sqlx.DB // 数据库连接
	MaxOpenCons int      // 最大连接数
	MaxIdleCons int      // 最大闲置连接
	DBPath      string   // sqlite数据库文件路径
}

func NewSqliteConner(connInfo SQLConnInfo) SqlConner {
	connInfo.DriverName = "sqlite3"
	db, err := sqlx.Open(connInfo.DriverName, connInfo.DBPath)
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

	return &SqliteInfo{
		DriverName:  connInfo.DriverName,
		DBConn:      connInfo.DBConn,
		MaxIdleCons: connInfo.MaxIdleCons,
		MaxOpenCons: connInfo.MaxOpenCons,
		DBPath:      connInfo.DBPath,
	}
}

func (connInfo *SqliteInfo) CreateTable() error {
	sql := "CREATE TABLE IF NOT EXISTS `schedule_jobs` (\n\t`id` INTEGER PRIMARY KEY AUTOINCREMENT," +
		"\n\t`cluster` VARCHAR(60) NULL DEFAULT NULL," +
		"\n\t`schedule_id` VARCHAR(60) NULL DEFAULT NULL," +
		"\n\t`job_id` VARCHAR(60) NULL DEFAULT NULL," +
		"\n\t`job_name` VARCHAR(60) NULL DEFAULT NULL," +
		"\n\t`job_account` VARCHAR(60) NULL DEFAULT NULL," +
		"\n\t`job_account_group` VARCHAR(60) NULL DEFAULT NULL," +
		"\n\t`job_queue` VARCHAR(60) NULL DEFAULT NULL," +
		"\n\t`job_status` VARCHAR(60) NULL DEFAULT NULL," +
		"\n\t`job_use_nodes` BIGINT NULL DEFAULT '0'," +
		"\n\t`job_node_list` TEXT NULL DEFAULT NULL," +
		"\n\t`job_use_cpus` BIGINT NULL DEFAULT '0'," +
		"\n\t`job_use_gpus` BIGINT NULL DEFAULT '0'," +
		"\n\t`job_exec_command` TEXT NULL DEFAULT NULL," +
		"\n\t`job_submit_time` VARCHAR(60) NULL DEFAULT NULL," +
		"\n\t`job_start_time` VARCHAR(60) NULL DEFAULT NULL," +
		"\n\t`job_end_time` VARCHAR(60) NULL DEFAULT NULL," +
		"\n\t`job_running_time` VARCHAR(60) NULL DEFAULT NULL," +
		"\n\t`job_work_dir` TEXT NULL DEFAULT NULL," +
		"\n\t`create_time` DATETIME NULL DEFAULT CURRENT_TIMESTAMP," +
		"\n\t`deleted_status` INT(11) NULL DEFAULT '0'" +
		"\n\t);"
	_, err := connInfo.DBConn.Exec(sql)
	if err != nil {
		return err
	}

	indexSql := "create unique index IF NOT EXISTS schedule_uindex" +
		"\n\ton schedule_jobs (schedule_id, job_id);"
	_, err = connInfo.DBConn.Exec(indexSql)
	if err != nil {
		return err
	}
	return nil
}

func (connInfo *SqliteInfo) InsertAndUpdateDataTable(Data interface{}) error {
	sql := "REPLACE INTO schedule_jobs(cluster,schedule_id," +
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
		":job_work_dir);"
	_, err := connInfo.DBConn.NamedExec(sql, Data.([]map[string]interface{}))
	if err != nil {
		return err
	}
	return nil
}
