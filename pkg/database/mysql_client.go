package database

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
)

//func main() {
//	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
//	dsn := "root:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
//	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
//	if err != nil {
//		os.Exit(1)
//	}
//	client :=  SQLClient{
//		Client: db,
//		Logger: nil,
//		Ctx:    nil,
//		Stop:   nil,
//	}
//	urls := client.GetRoots("mod2")
//	fmt.Sprintf("%v", urls)
//}

// NewRedisCache initializes a new cache
func NewMySQLClient(sqlOptions *SQLConfig, logger *log.Logger, ctx context.Context, stop chan os.Signal) (*SQLClient, error) {
	dsn := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8mb4&parseTime=True&loc=Local", sqlOptions.User, sqlOptions.Password, sqlOptions.DbUrl, sqlOptions.DbName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlclient := &SQLClient{
		Client: db,
		Logger: logger.WithFields(log.Fields{"cache": "redis"}),
		ctx:    ctx,
		stop:   stop,
	}
	//go rc.monitorWorker()
	return sqlclient, nil
}

type SQLConfig struct {
	DbUrl    string
	User     string
	Password string
	DbName	 string
}