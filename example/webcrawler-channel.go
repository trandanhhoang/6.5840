package main

import (
	"fmt"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

//TODO, crawl using parallel with thread.
// func Crawl(url string) {
// }
type Fetch struct{
	fetched map[string]bool 
}

func worker(url string,c chan []string){
	_, urls,err := fetcher.Fetch(url)
	if err!=nil{
		fmt.Println("not found: ",url)
		c <- []string{}
	}else{
		c <- urls
	}
}

func master(fetch *Fetch,c chan []string){
	counter := 1
	for urls := range c{
		for _,url:= range urls{
			if fetch.fetched[url] == true{
				continue
			}
			fetch.fetched[url] = true
			fmt.Println("found: ",url)

			counter++
			go worker(url,c)
		}

		counter--
		if counter == 0{
			break
		}
	}
}

func main() {
	fetch := &Fetch{
		fetched : make(map[string]bool),
	}

	c := make(chan []string)

	go func(){
		c <- []string{"https://golang.org/"}
	}()

	master(fetch, c)
}

func worker(){
	// TODO, just using fetcher.Fetch + return something for channel
	// parameter should be (url, channel)
	
}


func master(){
	// TODO, using for loop the CHANNEL, break when fetch all urls (hint to break: using counter, init = 1, after fetch counter-=1 -> counter == 0 -> return)
	// parameter should be (something mark fetched or not, channel)

	// hint:
	// for ... := range channel
	// call `worker` by `go routine`

}


func main(){
	// init something mark fetched or not
	// init the first "value" put to channel by: channel<-value using `go-routine`
	// call master

}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}
