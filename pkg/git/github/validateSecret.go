package github

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
)


func IsValidSignature(rawpayload []byte, r *http.Request, key string) bool {
	//X-Hub-Signature is the original header from github, but since this message is from echo we receive Webhook-secret
	gotHash := strings.SplitN(r.Header.Get("webhook-secret"), "=", 2)
	secretHeaderExists := true
	validKey := key != ""
	if gotHash[0] != "sha1" {
		secretHeaderExists = false
	}

	// If there's a key and no header or viceversa then fail, if there is no header and no key then pass
	if (validKey && !secretHeaderExists)  || (secretHeaderExists && !validKey){
		return false
	} else if !validKey && !secretHeaderExists {
		return true
	}

	hash := hmac.New(sha1.New, []byte(key))
	if _, err := hash.Write(rawpayload); err != nil {
		log.Printf("Cannot compute the HMAC for request: %s\n", err)
		return false
	}

	expectedHash := hex.EncodeToString(hash.Sum(nil))
	log.Println("EXPECTED HASH:", expectedHash)
	return gotHash[1] == expectedHash
}