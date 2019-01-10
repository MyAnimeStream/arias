package main

import (
	"github.com/myanimestream/arias/aria2"
	"log"
	"sync"
)

func test(client *aria2.Client) {
	log.Println("adding stuff")
	gid, err := client.AddUri([]string{
		"https://sample-videos.com/video123/mp4/720/big_buck_bunny_720p_2mb.mp4",
	}, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Waiting for download to complete")
	if err := gid.WaitForDownload(); err != nil {
		log.Fatal(err)
	}

	files, err := gid.GetFiles()
	if err != nil {
		log.Fatal()
	}

	for _, file := range files {
		log.Println(file.Path)
	}
}

func main() {
	log.Println("Hello world")
	client, err := aria2.NewClient("ws://localhost:6800/jsonrpc")
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := client.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		go func() {
			wg.Add(1)
			defer wg.Done()

			test(client)
		}()
	}

	wg.Wait()
}
