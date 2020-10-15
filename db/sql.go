package db

import (
	"BitrixInflux/libs"
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var db *gorm.DB

func SQL() *gorm.DB {
	return db.Debug()
}

func Connect(config *libs.DatabaseConfiguration) {
	var err error
	switch strings.ToLower(config.Driver) {
	case "mssql":
		db, err = gorm.Open("mssql", "sqlserver://"+config.Username+":"+config.Password+"@"+config.Host+":"+string(config.Port)+"?database="+config.Database)
	case "mysql", "mariadb":
		//log.Printf(config.Username+":"+config.Password+"@tcp("+config.Host+")/"+config.Database+"?charset=utf8&parseTime=True")
		db, err = gorm.Open("mysql", config.Username+":"+config.Password+"@tcp("+config.Host+")/"+config.Database+"?charset=utf8&parseTime=True&loc=Local")
		if err == nil {
			//db.Set("gorm:table_options", "ENGINE=MyISAM")
		}
	case "postgre", "postgres", "postgresql":
		pwd := ""
		if config.Password != "" {
			pwd = " password=" + config.Password
		}
		dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s"+pwd+" sslmode=disable", config.Host, config.Port, config.Username, config.Database)
		db, err = gorm.Open("postgres", dsn)
	case "sqlite", "sqlite3":
		db, err = gorm.Open("sqlite3", config.Database)
	}
	if err != nil {
		panic("Failed to connect database: ")
	}
}

func Close() (err error) {
	return db.Close()
}
