package crawler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"scraper/models"
	"scraper/redis_utils"
	"scraper/utils"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"github.com/gocolly/redisstorage"
	"github.com/joho/godotenv"
)

var UrlProcessCount int
var RedisTaskQueueName string
var RedisResultQueueName string
var RetriesLeft int
var CrawlerUserAgent string
var RedisHost string
var RedisPass string
var RedisPort string
var RedisGoCollyStoragePrefix string

func init() {
	_ = godotenv.Load(".env")
	UrlProcessCount, _ = strconv.Atoi(os.Getenv("URL_PROCESS_COUNT"))
	RedisTaskQueueName = os.Getenv("REDIS_TASK_QUEUE")
	RedisResultQueueName = os.Getenv("REDIS_RESULT_QUEUE")
	RetriesLeft, _ = strconv.Atoi(os.Getenv("RETRIES_LEFT"))
	CrawlerUserAgent = os.Getenv("CRAWLER_USER_AGENT")
	RedisHost = os.Getenv("REDIS_HOST")
	RedisPass = os.Getenv("REDIS_PASS")
	RedisPort = os.Getenv("REDIS_PORT")
	RedisGoCollyStoragePrefix = os.Getenv("REDIS_GO_COLLY_STORAGE_PREFIX")
	redis_utils.InitRedisMem()
	redis_utils.InitRedisBloom()
}

func Scrape(n int) {
	redis_utils.RedisConnect.RPush(RedisTaskQueueName, "")
	for {
		scrape()
		// time.Sleep(time.Second * 5)
		fmt.Println()
	}
}

func scrape() {

	dummy_host := ""
	u := redis_utils.RedisConnect.LPop(RedisTaskQueueName).(string)
	if u == "" {
		fmt.Println("No url to process...")
		return
	} else {
		fmt.Println("Processing.... ", u)
	}

	// Url process counter
	atomicCounter := utils.NewUrlProcessCount(UrlProcessCount)

	// parse hostname from url
	var hostname = utils.GetHostname(u)

	var c = colly.NewCollector(
		colly.AllowURLRevisit(),
	)
	c.UserAgent = CrawlerUserAgent

	c.Limit(&colly.LimitRule{
		RandomDelay: 2 * time.Second,
	})

	// create the redis storage
	storage := &redisstorage.Storage{
		Address:  RedisHost + ":" + RedisPort,
		Password: RedisPass,
		DB:       0,
		Prefix:   RedisGoCollyStoragePrefix + hostname,
	}

	// add redis in memory storage to the collector
	err := c.SetStorage(storage)
	if err != nil {
		panic(err)
	}

	// close redis client
	defer storage.Client.Close()

	// create a new request queue with redis storage backend
	q, _ := queue.New(5, storage)

	c.OnError(func(r *colly.Response, err error) {
		retriesLeft := RetriesLeft
		if x, ok := r.Ctx.GetAny("retriesLeft").(int); ok {
			retriesLeft = x
		}

		if retriesLeft > 0 {
			fmt.Printf("retry_attempt %d |  Error %s\n", retriesLeft, err.Error())
			r.Ctx.Put("retriesLeft", retriesLeft-1)
			time.Sleep((time.Duration(RetriesLeft) - time.Duration(retriesLeft) + 1) * time.Second)
			r.Request.Retry()
		}

	})

	c.OnResponse(func(r *colly.Response) {

		// Parsing the visited url
		content := Parser(r)
		if err != nil {
			fmt.Println(err.Error())
		}

		content.Domain = hostname
		content.URL = r.Request.URL.String()
		content.NormUrl = utils.NormalizeUrl(r.Request.URL)
		// err = content.InsertOne()

		jsonContent, err := json.Marshal(content)
		if err != nil {
			fmt.Println(err.Error())
		}

		// Pushing it to indexing queue
		redis_utils.RedisConnect.RPush(RedisResultQueueName, jsonContent)
	})

	c.OnHTML("a", func(e *colly.HTMLElement) {
		// Get the url of the element
		link := e.Request.AbsoluteURL(e.Attr("href"))

		if len(link) == 0 {
			return
		}

		urlHostname := utils.GetHostname(link)

		if urlHostname == "" {
			return
		}

		u, err := url.ParseRequestURI(link)

		if err != nil {
			fmt.Println("Invalid url : ", u.String())
			return
		}

		// Normalized url.
		link = utils.NormalizeUrl(u)

		// Check in bloom filter
		created, err := redis_utils.BloomFilterClient.Add(utils.Md5(link))

		if err != nil {
			fmt.Printf("error while adding the url to the crawler: %s\n", link)
		}

		if !created {
			// fmt.Printf("url :: %s is already queued to be crawled\n", u.String())
			return
		}

		// Prioritize links in the same domain
		if urlHostname != dummy_host {
			fmt.Printf("Found ext link %s\n", u.String())
		} else if !atomicCounter.Increase() {
			redis_utils.RedisConnect.RPush(RedisTaskQueueName, u.String())
		} else {
			q.AddURL(u.String())
		}
	})

	c.OnRequest(func(r *colly.Request) {

	})

	// add URLs to be processed
	q.AddURL(u)
	q.Run(c)

	fmt.Println(redis_utils.RedisConnect.LLen(RedisTaskQueueName), "this is the redis task queue list")
	fmt.Println(redis_utils.RedisConnect.LLen(RedisResultQueueName), "this is the result que list	")
}

func Parser(r *colly.Response) *models.Content {
	content := &models.Content{}

	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(r.Body))

	// Setting up title of the web page
	var title string
	title = doc.Find("title").Text()

	if len(title) == 0 {
		title = doc.Find("h1").Text()
	}

	if len(title) > 0 {
		content.Title = title
	}

	// Setting up the description of the web page
	var desc string

	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		if value, _ := s.Attr("name"); strings.Contains(value, "description") {
			desc, _ = s.Attr("content")
		}
		if len(desc) == 0 {
			if value, _ := s.Attr("property"); strings.Contains(value, "description") {
				desc, _ = s.Attr("content")
			}
		}
	})
	content.Desc = desc

	//TODO
	// Scrap the raw html and set up into the content model

	return content
}
