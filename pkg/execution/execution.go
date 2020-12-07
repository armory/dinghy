

package execution

import (
	"github.com/armory/dinghy/pkg/database"
	"time"
)

type Execution interface {
	Execute() (map[string]interface{}, error)
	CanExecute() bool
	CreateExecution(sqlClient *database.SQLClient, executionName string) error
	UpdateExecution(sqlClient *database.SQLClient, executionName string, result string, success bool) error
}

func CreateExecution (sqlClient *database.SQLClient, executionName string) error {
	nanos := time.Now().UnixNano()
	milis := nanos / 1000000
	execution := database.ExecutionSQL{
		Execution:       executionName,
		Result:          "",
		Success:         false,
		LastUpdatedDate: int(milis),
	}
	return sqlClient.Client.Create(&execution).Error
}

func UpdateExecution (sqlClient *database.SQLClient, executionName string, result string, success bool) error {
	nanos := time.Now().UnixNano()
	milis := nanos / 1000000
	execution := database.ExecutionSQL{
		Execution:       executionName,
		Result:          result,
		Success:         success,
		LastUpdatedDate: int(milis),
	}
	return sqlClient.Client.Save(&execution).Error
}