package util

import (
	"github.com/armory/plank/v4"
)

type PlankClient interface {
	GetApplication(string, string) (*plank.Application, error)
	UpdateApplicationNotifications(plank.NotificationsType, string, string) error
	GetApplicationNotifications(string, string) (*plank.NotificationsType, error)
	CreateApplication(*plank.Application, string) error
	UpdateApplication(plank.Application, string) error
	GetPipelines(string, string) ([]plank.Pipeline, error)
	DeletePipeline(plank.Pipeline, string) error
	UpsertPipeline(plank.Pipeline, string, string) error
	ResyncFiat(string) error
	ArmoryEndpointsEnabled() bool
	EnableArmoryEndpoints()
	UseGateEndpoints()
	UseServiceEndpoints()
}
