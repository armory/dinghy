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
	ctx    context.Context
	stop   chan os.Signal
}

type Fileurl struct {
	Id 			int		`gorm:"primaryKey;column:id"`
	Url			string	`gorm:"column:url"`
	Rawdata		string	`gorm:"column:rawdata"`
}

type FileurlChilds struct {
	FileurlID 			int		`gorm:"column:fileurl_id"`
	ChildfileurlId		int		`gorm:"column:childfileurl_id"`
}

// SetDeps sets dependencies for a parent
func (c *SQLClient) SetDeps(parent string, deps []string) {

	currParent := Fileurl{}
	c.Client.Where(&Fileurl{Url: parent}).Find(&currParent)
	if currParent.Url == "" {
		currParent.Url = parent
		c.Client.Create(&currParent)
	}
	children := []FileurlChilds{}
	c.Client.Where(&FileurlChilds{FileurlID: currParent.Id}).Find(&children)
	for _, currDep := range deps {
		foundDep := Fileurl{}
		c.Client.Where(&Fileurl{Url: currDep}).Find(&foundDep)
		if foundDep.Url == "" {
			foundDep.Url = currDep
			c.Client.Create(&foundDep)
		}
		if !containsChildId(children, foundDep.Id) {
			c.Client.Create(&FileurlChilds{
				FileurlID:      currParent.Id,
				ChildfileurlId: foundDep.Id,
			})
		}
	}

}

func containsChildId(slice []FileurlChilds, searchChildId int) bool {
	if slice == nil {
		return false
	}
	for _, val := range slice  {
		if val.ChildfileurlId == searchChildId {
			return true
		}
	}
	return false
}

// GetRoots grabs roots
func (c *SQLClient) GetRoots(url string) []string {
	return returnRoots(c, url)
}

func returnRoots (c *SQLClient, url string) []string {
	results := []string{}
	currUrl := Fileurl{}
	c.Client.Where(&Fileurl{Url: url}).Find(&currUrl)
	if currUrl.Url != "" {
		parents := []FileurlChilds{{FileurlID: currUrl.Id }}
		tempParent := []FileurlChilds{}
		for {
			for _, currParent := range parents {
				records := []FileurlChilds{}
				c.Client.Where(&FileurlChilds{ChildfileurlId: currParent.FileurlID}).Find(&records)
				if len(records) == 0 {
					resultUrl := Fileurl{}
					c.Client.Where(&Fileurl{Id: currParent.FileurlID}).Find(&resultUrl)
					results = append(results, resultUrl.Url)
				} else {
					for _, currRecord := range records {
						tempParent = append(tempParent, currRecord)
					}
				}
			}
			if len(tempParent) == 0 {
				break
			} else {
				parents = tempParent
				tempParent = []FileurlChilds{}
			}
		}
	}
	return results
}

// Set RawData
func (c *SQLClient) SetRawData(url string, rawData string) error {
	return c.Client.Model(&Fileurl{}).Where(&Fileurl{Url: url}).Update("rawdata", rawData).Error
}

func (c *SQLClient) GetRawData(url string) (string, error) {
	return returnRawData(c, url)
}

// Return RawData
func returnRawData( c *SQLClient, url string) (string, error) {
	find := Fileurl{Url: url}
	result := c.Client.Where(&Fileurl{Url: url}).Find(&find)
	return find.Rawdata, result.Error
}