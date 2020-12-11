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

package execution

import (
	"github.com/armory/dinghy/pkg/cache"
	"github.com/armory/dinghy/pkg/database"
	"github.com/armory/dinghy/pkg/settings"
	logr "github.com/sirupsen/logrus"
)

type RedisToSQLMigration struct {
	Settings *settings.Settings
	Logger *logr.Logger
	RedisCache *cache.RedisCache
	SQLClient *database.SQLClient
}

func (execution *RedisToSQLMigration) ExecutionName () string {
	return "REDIS_TO_SQL_MIGRATION"
}

func (execution *RedisToSQLMigration) CreateExecution(sqlClient *database.SQLClient, executionName string) error{
	return CreateExecution(sqlClient, executionName)
}

func (execution *RedisToSQLMigration) CanExecute() bool {
	if execution.RedisCache != nil && execution.Settings.SQL.Enabled && !execution.Settings.SQL.EventLogsOnly && execution.SQLClient != nil {
		if _, err := execution.RedisCache.Client.Ping().Result(); err != nil {
			return false
		}
		sqlDb, err := execution.SQLClient.Client.DB()
		if err == nil {
			if errPing := sqlDb.Ping(); errPing != nil {
				return false
			} else {
				found := database.ExecutionSQL{}
				result  := execution.SQLClient.Client.Where(&database.ExecutionSQL{Execution: execution.ExecutionName()}).Find(&found)
				if result.Error == nil && result.RowsAffected == 0 {
					return true
				}
			}
		}
	}
	return false
}

func (execution *RedisToSQLMigration) Finalize () {
	execution.Logger.Info("Closing Redis Client")
	err := execution.RedisCache.Client.Close()
	if err != nil {
		execution.Logger.Errorf("Failed to close redis client: %v", err)
	} else {
		execution.Logger.Info("Redis client was closed successfully")
	}
}

func (execution *RedisToSQLMigration) UpdateExecution (sqlClient *database.SQLClient, executionName string, result string, success bool) error {
	return UpdateExecution(sqlClient, executionName, result, success)
}

func (execution *RedisToSQLMigration) Execute() (map[string]interface{}, error){

	execution.Logger.Infof("Executing %v", execution.ExecutionName())

	if execution.CanExecute() == false {
		execution.Logger.Infof("%v will not be executed because CanExecute method returned false", execution.ExecutionName())
		return nil, nil
	}

	errorExec := CreateExecution(execution.SQLClient, execution.ExecutionName())
	if errorExec != nil {
		execution.Logger.Infof("%v will not be executed because %v", execution.ExecutionName(), errorExec)
		return nil, nil
	}

	files := execution.RedisCache.GetAllDinghyfiles()
	nextFiles := []string{}
	visited := map[string]bool{}
	for {

		for _, parentValue := range files {
			childrens := execution.RedisCache.GetChildren(parentValue)
			execution.SQLClient.SetDeps(parentValue, childrens)
			for _, currChildren := range childrens {
				if _, ok := visited[currChildren]; !ok {
					visited[currChildren] = true
					nextFiles = append(nextFiles, currChildren)
				}
			}
		}

		if len(nextFiles) == 0 {
			break
		}
		files = nextFiles
		nextFiles = []string{}
	}

	dinghyfiles := execution.RedisCache.GetAllDinghyfiles()
	for _, dinghyfile := range dinghyfiles {
		rawdata, err := execution.RedisCache.GetRawData(dinghyfile)
		if rawdata != "" && err == nil {
			execution.SQLClient.SetRawData(dinghyfile, rawdata)
		}
	}

	errorExec = UpdateExecution(execution.SQLClient, execution.ExecutionName(), "", true)
	if errorExec != nil {
		execution.Logger.Infof("%v was executed but execution registry failed to be updated because: %v", execution.ExecutionName(), errorExec)
		return nil, nil
	}
	execution.Logger.Infof("%v was successfully executed", execution.ExecutionName())
	return  nil, nil
}
