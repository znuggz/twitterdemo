package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/coreos/pkg/flagutil"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

func getTweets(ch chan int, w *csv.Writer, tweetMap map[string]int, mux *sync.Mutex, keepGoing bool) {

	if keepGoing == true {
		// randomly select a day during the last 7 days
		randDay := rand.Intn(7)
		currentTime := time.Now()
		dayInt := currentTime.Day() - randDay
		dayStr := strconv.Itoa(dayInt)
		if dayInt < 10 {
			dayStr = "0" + dayStr
		}

		// select a random language
		randLang := rand.Intn(5)
		lang := ""
		switch randLang {
		case 0:
			lang = "es"
		case 1:
			lang = "it"
		case 2:
			lang = "fr"
		case 3:
			lang = "pt"
		default:
			lang = "en"
		}

		// start twitter account access config
		flags := flag.NewFlagSet("user-auth", flag.ExitOnError)
		consumerKey := flags.String("consumer-key", "DInZ28FABMD4h5VQn8qRo1zD8", "Twitter Consumer Key")
		consumerSecret := flags.String("consumer-secret", "D2mtX6hPn42wYbCi1oTvNVOBKGaiS0N2YxYQADl56vReKjKmUt", "Twitter Consumer Secret")
		accessToken := flags.String("access-token", "31434194-UONEjSn8eMJ1m8RWum4R4TmYAy3oM4XCR7B7EFmYX", "Twitter Access Token")
		accessSecret := flags.String("access-secret", "6CLYyK12oGLb70241cuNN1Emsl6augPxkYpJde47xGJsS", "Twitter Access Secret")
		flags.Parse(os.Args[1:])
		flagutil.SetFlagsFromEnv(flags, "TWITTER")

		if *consumerKey == "" || *consumerSecret == "" || *accessToken == "" || *accessSecret == "" {
			log.Fatal("Consumer key/secret and Access token/secret required")
		}

		config := oauth1.NewConfig(*consumerKey, *consumerSecret)
		token := oauth1.NewToken(*accessToken, *accessSecret)

		// OAuth1 http.Client will automatically authorize Requests
		httpClient := config.Client(oauth1.NoContext, token)
		client := twitter.NewClient(httpClient)

		searchTweetParams := &twitter.SearchTweetParams{
			Query:      "#IoT",
			Count:      100,
			ResultType: "mixed",
			Lang:       lang,
			Until:      "2019-05-" + dayStr,
		}

		search, _, err := client.Search.Tweets(searchTweetParams)
		if err != nil {
			fmt.Println(err)
		}

		for i := 0; i < len(search.Statuses); i++ {
			mux.Lock()
			_, ok := tweetMap[search.Statuses[i].IDStr]
			if ok == false {
				tweetMap[search.Statuses[i].IDStr] = 1
				// function that writes to file here
				v := reflect.ValueOf(search.Statuses[i])
				values := make([]interface{}, v.NumField())
				stringSlice := make([]string, 20)
				for i := 0; i < v.NumField(); i++ {
					values[i] = v.Field(i).Interface()
					stringSlice = append(stringSlice, fmt.Sprintf("%v", values[i]))
				}

				if err := w.Write(stringSlice); err != nil {
					log.Fatalln("error writing record to csv:", err)
				}
			}

			// Write any buffered data to the underlying writer (standard output).
			w.Flush()

			if err := w.Error(); err != nil {
				log.Fatal(err)
			}
			mux.Unlock()
		}

		ch <- len(tweetMap)
	} else {
		close(ch)
	}
}

func main() {
	// file handle and error check
	f, err := os.Create("/tmp/tweets.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	// to check uniqueness
	tweetMap := make(map[string]int)
	// to lock for writes to file and map
	mux := sync.Mutex{}
	ch := make(chan int)

	requestCounter := 0
	numThreds := 10
	tweetCounter := 0
	for tweetCounter <= 2000 {
		// spin up 10 threads
		for requests := 0; requests < numThreds; requests++ {
			go getTweets(ch, w, tweetMap, &mux, true)
			requestCounter++
		}
		fmt.Printf("how many threads: %d\n", requestCounter)

		// check the returns
		for i := 0; i < numThreds; i++ {
			current, ok := <-ch
			fmt.Printf("keep going? %t\n", ok)
			fmt.Printf("current: %d\n", current)
			if current > tweetCounter {
				tweetCounter = current
				fmt.Printf("how many tweets? %d\n", tweetCounter)
			}
		}

		// if we have over 2000 tweets close the channel
		if tweetCounter >= 2000 {
			go getTweets(ch, w, tweetMap, &mux, false)
		} else {
			// we have hit tweeter's max request limit, need to wait for another window
			if requestCounter == 180 {
				fmt.Println("time to take a 15 minute nap...")
				time.Sleep(5 * time.Minute)
				fmt.Println("time to take a 10 minute nap...")
				time.Sleep(5 * time.Minute)
				fmt.Println("time to take a 5 minute nap...")
				time.Sleep(5 * time.Minute)
				requestCounter = 0
			}
		}
	}
	fmt.Println("done!")
}
