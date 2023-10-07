package dal

import (
	"fmt"
	"time"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

const dsnTemplate = "%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local"

func MustInitMySQL() {
	var err error

	c := conf.Conf
	dsn := fmt.Sprintf(dsnTemplate, c.MySQL.Username, c.MySQL.Password, c.MySQL.Host, c.MySQL.Port, c.MySQL.DBName, c.MySQL.Charset)

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("fatal error db init: %w", err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if c.Base.IsDebug {
		db = db.Debug()
	}

	fmt.Println("mysql initialized")
}

func DB() *gorm.DB {
	return db
}

func GetQuery() *query.Query {
	return query.Use(db)
}
