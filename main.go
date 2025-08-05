package main

import (
	"flag"
	"fmt"
	"scraper/crawler"
)

var workerMap map[string]func(count int) = map[string]func(count int){
	"worker":  crawler.RunIndexer,
	"crawler": crawler.Scrape,
}

func main() {
	var worker string
	var count int

	flag.StringVar(&worker, "worker", "crawler", "a string")
	flag.IntVar(&count, "count", 1, "count of workers")
	flag.Parse()

	if _, ok := workerMap[worker]; !ok {
		fmt.Println("Invalid worker type defined.")
		return
	}

	if count <= 0 {
		fmt.Println("invalid count argument")
		return
	}

	f := workerMap[worker]
	f(count)
}
