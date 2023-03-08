package dal

import (
	"fmt"

	"kp-management/internal/pkg/conf"
	"kp-management/internal/pkg/dal/query"

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
