package mysql

import (
	"bluebell/settings"
	"fmt"

	"go.uber.org/zap"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

func Init(cfg *settings.MySQLConfig) (err error) {
	var dbConfig string
	dbConfig = getMysqlConfig(cfg)

	db, err = gorm.Open("mysql", dbConfig)
	if err == nil {
		//fmt.Printf("mysql init failed %v\n", err)
		zap.L().Error("grom init failed %v\n", zap.Error(err))
	}

	//GORM定义mysql最大连接数
	sqlDB := db.DB()                        //设置数据库连接池参数
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns) //设置数据库连接池最大连接数
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns) //连接池最大允许的空闲连接数，如果没有sql任务需要执行的连接数大于20，超过的连接会被连接池关闭。

	// 禁用默认表名的复数形式，如果置为 true，则 `User` 的默认表名是 `user`
	db.SingularTable(true)
	return
}

func getMysqlConfig(cfg *settings.MySQLConfig) string {
	return fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.DB)
}

func Close() (err error) {
	err = db.Close()
	if err != nil {
		//fmt.Printf("grom close failed error:%v\n", err)
		zap.L().Error("grom close failed error:%v\n", zap.Error(err))
	}
	return
}
