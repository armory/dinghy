package util

import (
	"github.com/armory/plank"
)

type PlankClient interface {
	GetApplication(string) (*plank.Application, error)
	CreateApplication(*plank.Application) error
	GetPipelines(string) ([]plank.Pipeline, error)
	DeletePipeline(plank.Pipeline) error
	UpsertPipeline(plank.Pipeline, string) error
	ResyncFiat() error
	ArmoryEndpointsEnabled() bool
	EnableArmoryEndpoints()
}
