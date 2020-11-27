/*
* Copyright 2019 Armory, Inc.

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

package main

import (
	"fmt"
	"github.com/armory/dinghy/pkg/database"
	//"github.com/armory/dinghy/cmd"
	//"github.com/armory/dinghy/pkg/settings"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
)

//func main() {
//	log, d := dinghy.Setup()
//	config := settings.NewDefaultSettings()
//	dinghy.Start(log, d, &config)
//}


func main() {
	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	dsn := "root:password@tcp(127.0.0.1:3306)/dinghy?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		os.Exit(1)
	}
	client :=  database.SQLClient{
		Client: db,
		Logger: nil,
		Ctx:    nil,
		Stop:   nil,
	}
	urls := client.GetRoots("mod2")
	fmt.Printf("%v", urls)

	rawdata,err := client.GetRawData("test2")
	fmt.Printf("%v", rawdata)

	errinsert := client.SetRawData("test3", "rawdata4")
	fmt.Printf("%v", errinsert)
	rawdata3,err := client.GetRawData("test3")
	fmt.Printf("%v", rawdata3)
}