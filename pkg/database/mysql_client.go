package database

import (
"fmt"
"gorm.io/driver/mysql"
"gorm.io/gorm"
"os"
)

func main() {
	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	dsn := "root:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		os.Exit(1)
	}
	client :=  SQLClient{
		Client: db,
		Logger: nil,
		Ctx:    nil,
		Stop:   nil,
	}
	urls := client.GetRoots("mod2")
	fmt.Sprintf("%v", urls)
}