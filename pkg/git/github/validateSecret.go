package github

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	log "github.com/sirupsen/logrus"
	"strings"
)


func IsValidSignature(rawpayload []byte, webhookSecret string, key string) bool {
	gotHash := strings.SplitN(webhookSecret, "=", 2)
	if gotHash[0] != "sha1" {
		log.Error("Invalid webhook value")
	}

	hash := hmac.New(sha1.New, []byte(key))
	if _, err := hash.Write(rawpayload); err != nil {
		log.Printf("Cannot compute the HMAC for request: %s\n", err)
		return false
	}

	expectedHash := hex.EncodeToString(hash.Sum(nil))
	log.Printf("Expected hash: %v and got %v from github", expectedHash, gotHash[1])
	return gotHash[1] == expectedHash
}