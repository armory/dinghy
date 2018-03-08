package main

import (
	"github.com/armory-io/dinghy/pkg/web"
	"github.com/armory-io/dinghy/pkg/cache"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	log.SetLevel(log.DebugLevel)
	log.Info("Dinghy started.")
	cache.C = cache.NewCache()
	log.Fatal(http.ListenAndServe(":8080", web.Router()))
}
