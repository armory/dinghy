package web

import (
	dinghylog "github.com/armory/dinghy/pkg/log"
	"github.com/dlclark/regexp2"
	"regexp"
)

type IgnoreFile interface {
	ShouldIgnore(filename string) bool
}

type RegexpIgnoreFile struct {
	logger   dinghylog.DinghyLog
	patterns []string
}

type Regexp2IgnoreFile struct {
	logger  dinghylog.DinghyLog
	regExps []*regexp2.Regexp
}

func (t *RegexpIgnoreFile) ShouldIgnore(filename string) bool {
	for _, pattern := range t.patterns {
		if result, _ := regexp.MatchString(pattern, filename); result {
			t.logger.Infof("file %s matches pattern %s: %t", filename, pattern, result)
			return true
		}
	}
	return false
}

func (t *Regexp2IgnoreFile) ShouldIgnore(filename string) bool {
	for _, regExp := range t.regExps {
		if result, _ := regExp.MatchString(filename); result {
			t.logger.Infof("file %s matches pattern %s: %t", filename, regExp.String(), result)
			return true
		}
	}
	return false
}

func NewRegexpIgnoreFile(patterns []string, logger dinghylog.DinghyLog) IgnoreFile {
	return &RegexpIgnoreFile{
		logger:   logger,
		patterns: patterns,
	}
}

func NewRegexp2IgnoreFile(patterns []string, logger dinghylog.DinghyLog) IgnoreFile {
	var regExps []*regexp2.Regexp
	for _, pattern := range patterns {
		regExps = append(regExps, regexp2.MustCompile(pattern, 0))
	}
	return &Regexp2IgnoreFile{
		logger:  logger,
		regExps: regExps,
	}
}
