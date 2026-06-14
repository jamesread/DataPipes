package main

import "github.com/jamesread/data-cleaner/internal/httpservers"
import "github.com/jamesread/data-cleaner/internal/config"
import log "github.com/sirupsen/logrus"

func main() {
	log.Infof("DataPipes")

	config := config.GetConfig()

	log.Infof("config %+v", config)

	httpservers.Start()
}
