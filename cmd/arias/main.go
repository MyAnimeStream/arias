package main

import (
	"flag"
	"github.com/MyAnimeStream/arias"
	"log"
)

func main() {
	configFile := flag.String("config", "", "Specify config path")
	flag.Parse()

	config, err := arias.LoadConfig(*configFile)
	if err != nil {
		log.Fatal("Error loading config: ", err)
	}

	if err = config.Check(); err != nil {
		log.Fatal("Config error: ", err)
	}

	server, err := arias.NewServer(config)
	if err != nil {
		log.Fatal("Couldn't start server: ", err)
	}

	log.Fatal(server.ListenAndServe())
}
