package crawler

import (
	"encoding/json"
	"fmt"
	"os"
	"scraper/models"
	"scraper/redis_utils"
	"strconv"
	"sync"
	"time"
)

var MongoBatchInsertionSize int

func init() {
	redis_utils.InitRedisMem()
	redis_utils.InitRedisBloom()
	MongoBatchInsertionSize, _ = strconv.Atoi(os.Getenv("MONGO_BATCH_INSERTION_SIZE"))

}

type Indexer struct {
	jobs chan *models.Content
	wg   *sync.WaitGroup
}

func NewIndexer() *Indexer {
	return &Indexer{
		jobs: make(chan *models.Content, 1000000),
		wg:   &sync.WaitGroup{},
	}
}

func (indexer *Indexer) addJob(job *models.Content) {
	indexer.jobs <- job
}

func (indexer *Indexer) indexToMongo(jobs []any) {
	retries := 3

	for retries > 0 {
		err := models.InsertMany(jobs)

		if err != nil {
			fmt.Printf("Error while indexing to the mongo db : %s\n", err.Error())
			retries--
		} else {
			return
		}
	}
}

func (indexer *Indexer) initiateWorkers(n int) {
	if n < 0 {
		n = 1
	}

	indexer.wg.Add(1)

	for range n {
		go func() {
			ticker := time.NewTicker(time.Second * 50)
			jobData := make([]any, 0)
			for {
				select {
				case <-ticker.C:
					fmt.Printf("syncing to monog db %d", len(jobData))
					if len(jobData) > 0 {
						indexer.indexToMongo(jobData)
						jobData = make([]any, 0)
					}
				case job := <-indexer.jobs:
					jobData = append(jobData, job)

					if len(jobData) > int(MongoBatchInsertionSize) {
						indexer.indexToMongo(jobData)
						jobData = make([]any, 0)
					}
				}
			}
		}()
	}
	indexer.wg.Wait()
}

func RunIndexer(n int) {

	mainIndexer := NewIndexer()
	go func() {
		for {
			if len(mainIndexer.jobs) > 10000 {
				continue
			}
			val := redis_utils.RedisConnect.LPop(RedisResultQueueName)

			if val != "" {
				content := &models.Content{}
				err := json.Unmarshal([]byte(val.(string)), content)

				if err != nil {
					fmt.Println("invalid content to unmarshal...")
					continue
				}
				mainIndexer.addJob(content)
			}
		}
	}()
	mainIndexer.initiateWorkers(n)
}
