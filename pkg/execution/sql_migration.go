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


func (execution *RedisToSQLMigration) CreateExecution(sqlClient *database.SQLClient, executionName string) error{
	return CreateExecution(sqlClient, executionName)
}

func (execution *RedisToSQLMigration) CanExecute() bool {
	if execution.RedisCache != nil && execution.Logger != nil && execution.Settings.SQL.Enabled && !execution.Settings.SQL.EventLogsOnly && execution.SQLClient != nil {
		if _, err := execution.RedisCache.Client.Ping().Result(); err != nil {
			return false
		}
		sqlDb, err := execution.SQLClient.Client.DB()
		if err == nil {
			if errPing := sqlDb.Ping(); errPing != nil {
				return false
			} else {
				found := database.ExecutionSQL{}
				result  := execution.SQLClient.Client.Where(&database.ExecutionSQL{Execution: "REDIS_TO_SQL_MIGRATION"}).Find(&found)
				if result.Error == nil && result.RowsAffected == 0 {
					return true
				}
			}
		}
	}
	return false
}

func (execution *RedisToSQLMigration) UpdateExecution (sqlClient *database.SQLClient, executionName string, result string, success bool) error {
	return UpdateExecution(sqlClient, executionName, result, success)
}

func (execution *RedisToSQLMigration) Execute() (map[string]interface{}, error){

	if execution.CanExecute() == false {
		execution.Logger.Info("REDIS_TO_SQL_MIGRATION will not be executed because CanExecute method returned false")
		return nil, nil
	}

	errorExec := CreateExecution(execution.SQLClient, "REDIS_TO_SQL_MIGRATION")
	if errorExec != nil {
		execution.Logger.Info("REDIS_TO_SQL_MIGRATION will not be executed because %v", errorExec)
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

	errorExec = UpdateExecution(execution.SQLClient, "REDIS_TO_SQL_MIGRATION", "", true)
	if errorExec != nil {
		execution.Logger.Info("REDIS_TO_SQL_MIGRATION was executed but execution registry failed to be updated because: ", errorExec)
		return nil, nil
	}
	return  nil, nil
}
