package Dao

import "github.com/jmoiron/sqlx"

type SQLConnInfo struct {
	DriverName  string   //驱动名称
	User        string   // 连接用户名
	Password    string   // 连接密码
	Host        string   // mysql服务器地址
	Port        int      // mysql服务端口号
	DbName      string   // 数据库名
	DBConn      *sqlx.DB // 数据库连接
	MaxOpenCons int      // 最大连接数
	MaxIdleCons int      // 最大闲置连接
	DBType      int      // 数据库类型
	DBPath      string   // sqlite数据库文件路径
}

type SqlConner interface {
	CreateTable() error
	InsertAndUpdateDataTable(interface{}) error
}
