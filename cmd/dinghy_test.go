package dinghy

import (
	"github.com/armory/dinghy/pkg/dinghyfile"
	"github.com/armory/dinghy/pkg/util"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestDebugLevel(t *testing.T) {
	os.Setenv("DEBUG_LEVEL", "DEBUG")
	util.CopyToLocalSpinnaker("testdata/dinghy-local.yml", "dinghy-local.yml")

	logger, _ := Setup()

	assert.Equal(t, "debug", logger.Level.String())
}

func TestParserFormat(t *testing.T) {
	util.CopyToLocalSpinnaker("testdata/dinghy-local.yml", "dinghy-local.yml")

	_, webApi := Setup()

	assert.Equal(t, dinghyfile.NewDinghyfileParser(&dinghyfile.PipelineBuilder{}), webApi.Parser)
}
