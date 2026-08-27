package config

import (
	"fmt"
	"log"
	"megin/pkg/database"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 管理多库连接,读写分离或多库,微服务中不建议连接多个库，如果需要的话可以在这里扩展
type DBManager struct {
	database    *gorm.DB
	RedisClient *database.RedisClient
}

var dbManager = new(DBManager)

func InitDatabase(conf *ServiceConfig) {
	//连接mysql
	if conf.Database.Driver == "mysql" && len(conf.Database.Dsn) > 0 {
		db, err := gorm.Open(mysql.Open(conf.Database.Dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})

		if err != nil {
			log.Fatalln("连接Mysql数据库失败" + err.Error())
		} else {
			dbManager.database = db
		}
	}

	//连接Redis
	dbManager.RedisClient = database.RedisConnect(conf.Redis.Addr, conf.Redis.Password)
}

// 获取mysql库
func GetMysqlDB() *gorm.DB {
	if dbManager.database == nil {
		fmt.Println("mysql nil")
	}
	return dbManager.database
}

// 获取Redis
func GetRedis() *database.RedisClient {
	return dbManager.RedisClient
}
