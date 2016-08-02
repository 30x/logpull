package server

import (
  "log"
  "os"
  "strconv"
)

const (
  // DefaultPort is the default port to listen
  DefaultPort = "8000"
  // DefaultHitLimit is the default limit on the number of hits that can be pulled at one time
  DefaultHitLimit = 1024
)

// Port the port the server is listening on
var Port string
// ElasticSearchHost is the host name of the elastic search pod
var ElasticSearchHost string
// ElasticSearchPort  is the port of the elastic search pod
var ElasticSearchPort string
// HitLimit the limit on the number of hits that can be pulled at one time
var HitLimit int

// ConfigureLogPull configures the logpull server from environment variables
func ConfigureLogPull() {
  if Port = os.Getenv("PORT"); Port == "" {
    Port = DefaultPort
  }

  if HitLimitStr := os.Getenv("HIT_LIMIT"); HitLimitStr == "" {
    HitLimit = DefaultHitLimit
  } else {
    var err error

    HitLimit, err = strconv.Atoi(HitLimitStr)
    if err != nil {
      HitLimit = DefaultHitLimit
    }
  }

  if ElasticSearchHost = os.Getenv("ELASTIC_SEARCH_HOST"); ElasticSearchHost == "" {
    log.Fatal("Missing required variable ELASTIC_SEARCH_HOST")
  }

  ElasticSearchPort = os.Getenv("ELASTIC_SEARCH_PORT")
}