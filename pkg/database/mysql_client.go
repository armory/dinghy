package database

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"syscall"
	"time"
)

// NewMySQLClient initializes a MySQL Client
func NewMySQLClient(sqlOptions *SQLConfig, logger *log.Logger, ctx context.Context, stop chan os.Signal) (*SQLClient, error) {
	dsn := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8mb4&parseTime=True&loc=Local", sqlOptions.User, sqlOptions.Password, sqlOptions.DbUrl, sqlOptions.DbName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		return nil, err
	}
	sqlclient := &SQLClient{
		Client: db,
		Logger: logger.WithFields(log.Fields{"persistence": "sql"}),
		ctx:    ctx,
		stop:   stop,
	}

	go sqlclient.monitorWorker()
	return sqlclient, nil
}

func (c *SQLClient) monitorWorker() {
	timer := time.NewTicker(10 * time.Second)
	count := 0
	for {
		select {
		case <-timer.C:
			sqlDB, err := c.Client.DB()
			if err != nil {
				c.stop <- syscall.SIGINT
			}
			if err := sqlDB.Ping(); err != nil {
				count++
				c.Logger.Errorf("SQL monitor failed %d times (5 max)", count)
				if count >= 5 {
					c.Logger.Error("Stopping dinghy because communication with MySQL database failed")
					timer.Stop()
					c.stop <- syscall.SIGINT
				}
				continue
			}
			count = 0
		case <-c.ctx.Done():
			return
		}
	}
}

type SQLConfig struct {
	DbUrl    string
	User     string
	Password string
	DbName	 string
}