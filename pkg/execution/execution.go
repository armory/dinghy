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
	"github.com/armory/dinghy/pkg/database"
	"time"
)

type Execution interface {
	Execute() (map[string]interface{}, error)
	CanExecute() bool
	ExecutionName() string
	CreateExecution(sqlClient *database.SQLClient, executionName string) error
	UpdateExecution(sqlClient *database.SQLClient, executionName string, result string, success bool) error
	Finalize()
}

func CreateExecution(sqlClient *database.SQLClient, executionName string) error {
	nanos := time.Now().UnixNano()
	milis := nanos / 1000000
	execution := database.ExecutionSQL{
		Execution:       executionName,
		Result:          "",
		Success:         "false",
		LastUpdatedDate: int(milis),
	}
	return sqlClient.Client.Create(&execution).Error
}

func UpdateExecution(sqlClient *database.SQLClient, executionName string, result string, success bool) error {
	nanos := time.Now().UnixNano()
	milis := nanos / 1000000
	succVal := "false"
	if success {
		succVal = "true"
	}
	execution := database.ExecutionSQL{
		Execution:       executionName,
		Result:          result,
		Success:         succVal,
		LastUpdatedDate: int(milis),
	}
	return sqlClient.Client.Save(&execution).Error
}
