package database

import (
	"context"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"os"
)

type SQLClient struct {
	Client *gorm.DB
	Logger *log.Entry
	Ctx    context.Context
	Stop   chan os.Signal
}

type Dependency struct {
	Id 			int		`gorm:"primaryKey;column:id"`
	Url			string	`gorm:"column:url"`
	Parentid	int		`gorm:"column:parentid"`
}

type Rawdata struct {
	Url 		string	`gorm:"primaryKey;column:url"`
	Rawdata		string	`gorm:"column:rawdata"`
}

// SetDeps sets dependencies for a parent
func (c *SQLClient) SetDeps(parent string, deps []string) {
	//var getcurrent Dependency
	//
	//var childs []Dependency
	//c.Client.Select(&childs,  )
}

// GetRoots grabs roots
func (c *SQLClient) GetRoots(url string) []string {
	results := []string{}
	reference := []Dependency{}
	parentIds := []int{}
	c.Client.Where(&Dependency{Url: url}).Find(&reference)
	for _, currentRef := range reference {
		parentIds = append(parentIds, currentRef.Parentid)
	}
	if len(parentIds) > 0 {
		for {
			c.Client.Where("id IN ?", parentIds).Find(&reference)
			parentIds = []int{}
			if len(reference) == 0 {
				break
			}
			for _, val := range reference {
				if val.Parentid == 0 {
					results = append(results, val.Url)
				} else {
					parentIds = append(parentIds, val.Parentid)
				}
			}
		}
	}
	return results
}

// Set RawData
func (c *SQLClient) SetRawData(url string, rawData string) error{
	currentRawData, err := c.GetRawData(url)
	if err != nil {
		return err
	}
	if currentRawData != "" {
		return c.Client.Model(&Rawdata{Url: url}).Update("rawdata", rawData).Error
	}
	return c.Client.Create(&Rawdata{
		Url:     url,
		Rawdata: rawData,
	}).Error
}

// Return RawData
func (c *SQLClient) GetRawData(url string) (string, error) {
	find := Rawdata{Url: url}
	result := c.Client.Where(&Rawdata{Url: url}).Find(&find)
	return find.Rawdata, result.Error
}