package util

import (
	"github.com/armory/plank/v3"
)

type PlankClient interface {
	GetApplication(string) (*plank.Application, error)
	UpdateApplicationNotifications(plank.NotificationsType, string) error
	GetApplicationNotifications(string) (*plank.NotificationsType, error)
	CreateApplication(*plank.Application) error
	UpdateApplication(plank.Application) error
	GetPipelines(string) ([]plank.Pipeline, error)
	DeletePipeline(plank.Pipeline) error
	UpsertPipeline(plank.Pipeline, string) error
	ResyncFiat() error
	ArmoryEndpointsEnabled() bool
	EnableArmoryEndpoints()
}
