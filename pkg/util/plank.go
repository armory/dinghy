package util

import (
	"github.com/armory/plank"
)

type PlankClient interface {
	GetApplication(string) (*plank.Application, error)
	CreateApplication(*plank.Application) error
	GetPipeline(app, pipeline string) (*plank.Pipeline, error)
	GetPipelines(string) ([]plank.Pipeline, error)
	DeletePipeline(plank.Pipeline) error
	UpsertPipeline(plank.Pipeline) error
	WithFiatUser(string) PlankClient
}
