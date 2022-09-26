package web

import (
	dinghylog "github.com/armory/dinghy/pkg/log"
	"github.com/armory/dinghy/pkg/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_RegexpIgnoreFile_ShouldIgnore(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	l := mock.NewMockFieldLogger(c)
	l.EXPECT().Infof(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)

	dl := dinghylog.NewDinghyLogs(l)

	ignoreFile := NewRegexpIgnoreFile([]string{"file.(js|ts|css)"}, dl)

	assert.True(t, ignoreFile.ShouldIgnore("file.js"))
	assert.True(t, ignoreFile.ShouldIgnore("file.ts"))
	assert.True(t, ignoreFile.ShouldIgnore("file.css"))
	assert.False(t, ignoreFile.ShouldIgnore("dinghyfile"))
	assert.False(t, ignoreFile.ShouldIgnore("minimum-wait.stage.module"))
	assert.False(t, ignoreFile.ShouldIgnore("maximum-wait.stage.module"))
}

func Test_RegexpIgnoreFile_ShouldIgnoreWhenNegativeExpression(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	l := mock.NewMockFieldLogger(c)

	dl := dinghylog.NewDinghyLogs(l)

	ignoreFile := NewRegexpIgnoreFile([]string{"^(?!.*(.stage.module)|(dinghyfile)).*"}, dl)

	assert.False(t, ignoreFile.ShouldIgnore("file.js"))
	assert.False(t, ignoreFile.ShouldIgnore("file.ts"))
	assert.False(t, ignoreFile.ShouldIgnore("file.css"))
	assert.False(t, ignoreFile.ShouldIgnore("dinghyfile"))
	assert.False(t, ignoreFile.ShouldIgnore("minimum-wait.stage.module"))
	assert.False(t, ignoreFile.ShouldIgnore("maximum-wait.stage.module"))
}

func Test_Regexp2IgnoreFile_ShouldIgnore(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	l := mock.NewMockFieldLogger(c)
	l.EXPECT().Infof(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)

	dl := dinghylog.NewDinghyLogs(l)

	ignoreFile := NewRegexp2IgnoreFile([]string{"file.(js|ts|css)"}, dl)

	assert.True(t, ignoreFile.ShouldIgnore("file.js"))
	assert.True(t, ignoreFile.ShouldIgnore("file.ts"))
	assert.True(t, ignoreFile.ShouldIgnore("file.css"))
	assert.False(t, ignoreFile.ShouldIgnore("dinghyfile"))
	assert.False(t, ignoreFile.ShouldIgnore("minimum-wait.stage.module"))
	assert.False(t, ignoreFile.ShouldIgnore("maximum-wait.stage.module"))
}

func Test_Regexp2IgnoreFile_ShouldIgnoreWhenNegativeExpression(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	l := mock.NewMockFieldLogger(c)
	l.EXPECT().Infof(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(3)

	dl := dinghylog.NewDinghyLogs(l)

	ignoreFile := NewRegexp2IgnoreFile([]string{"^(?!.*(.stage.module)|(dinghyfile)).*"}, dl)

	assert.True(t, ignoreFile.ShouldIgnore("file.js"))
	assert.True(t, ignoreFile.ShouldIgnore("file.ts"))
	assert.True(t, ignoreFile.ShouldIgnore("file.css"))
	assert.False(t, ignoreFile.ShouldIgnore("dinghyfile"))
	assert.False(t, ignoreFile.ShouldIgnore("minimum-wait.stage.module"))
	assert.False(t, ignoreFile.ShouldIgnore("maximum-wait.stage.module"))
}
