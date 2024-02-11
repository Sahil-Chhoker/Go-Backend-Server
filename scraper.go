package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/sahil-chhoker/Go-Backend-Server/internal/database"
)

func startScraping(db *database.Queries, concurrency int, timeBetweenRequest time.Duration) {
	log.Printf("Scraping on %v goroutines every %s duration", concurrency, timeBetweenRequest)
	ticker := time.NewTicker(timeBetweenRequest)

	for ; ; <-ticker.C {
		feeds, err := db.GetNextFeedsToFetch(context.Background(), int32(concurrency))
		if err != nil {
			log.Println("error fetching feeds: ", err)
			continue
		}

		wg := &sync.WaitGroup{}
		for _, feed := range feeds {
			wg.Add(1)

			go scrapeFeed(db, wg, feed)
		}
		wg.Wait()
	}
}

func scrapeFeed(db *database.Queries, wg *sync.WaitGroup, feed database.Feed) {
	defer wg.Done()

	_, err := db.MarkFeedsAsFetched(context.Background(), feed.ID)
	if err != nil {
		log.Println("error marking feed ad fetched :", err)
		return
	}

	rssFeed, err := urlToFeed(feed.Url)
	if err != nil {
		log.Println("error fetching feed :", err)
		return
	}

	for _, item := range rssFeed.Channel.Item {
		log.Println("Found post ", item.Title, " on feed ", feed.Name)
	}
	log.Printf("Feed %s collected, %v postes found", feed.Name, len(rssFeed.Channel.Item))
}
