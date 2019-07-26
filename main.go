package main

import "log"

func main() {
	capacity := 10
	cache := New(capacity)

	start := "1"

	log.Printf("Cache with in-memory capacity: %d created.\n", capacity)

	for i := 0; i < 15; i++ {
		start = start + "1"
		cache.Set(start, start)
		log.Printf("Added one more key. Using disk: %v.\n", cache.usingDisk)
	}

}
