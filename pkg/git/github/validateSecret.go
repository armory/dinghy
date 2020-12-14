package github

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"github.com/armory/dinghy/pkg/log"
	"strings"
)

func IsValidSignature(rawpayload []byte, webhookSecret string, key string, logger log.DinghyLog) bool {
	gotHash := strings.SplitN(webhookSecret, "=", 2)
	if gotHash[0] != "sha1" {
		logger.Error("Invalid webhook value")
	}

	hash := hmac.New(sha1.New, []byte(key))
	if _, err := hash.Write(rawpayload); err != nil {
		logger.Printf("Cannot compute the HMAC for request: %s\n", err)
		return false
	}

	expectedHash := hex.EncodeToString(hash.Sum(nil))
	validation := gotHash[1] == expectedHash
	logger.Printf("Result from hash validation was: %v", validation)
	if !validation {
		logger.Error("Invalid webhook secret signature")
	}
	return validation
}
