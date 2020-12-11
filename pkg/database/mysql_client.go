/*
* Copyright 2020 Armory, Inc.
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*    http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

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

	//TODO: Add logger so queries can be seen, create a parameter later for this
	sqlclient := &SQLClient{
		Client: db,
		Logger: nil,
		ctx:    ctx,
		stop:   stop,
	}

	go sqlclient.monitorWorker()
	return sqlclient, nil
}

func (c *SQLClient) monitorWorker() {
	logger := log.WithFields(log.Fields{"persistence": "sql"})
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
				logger.Errorf("SQL monitor failed %d times (5 max)", count)
				if count >= 5 {
					logger.Error("Stopping dinghy because communication with MySQL database failed")
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
	DbName   string
}
