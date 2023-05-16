package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	Fetch(url string) (body string, urls []string, err error)
}

type Fetch struct{
	crawledUrl map[string]bool
	mu sync.Mutex
}

func Craw(url string, fetch *Fetch){
	fetch.mu.Lock()
	if fetch.crawledUrl[url] == true{
		defer fetch.mu.Unlock()
		return
	}
	fetch.crawledUrl[url] = true
	fetch.mu.Unlock()
	fmt.Println("found: ",url)

	var done sync.WaitGroup

	_,urls,err := fetcher.Fetch(url)
	if err!=nil{
		fmt.Println("err",err)
		return
	}

	for _,v := range urls{
		done.Add(1)
		go func(v string){
			defer done.Done()
			Craw(v,fetch)
		}(v)
	}
	
	done.Wait()
}

func main() {
	fetch := &Fetch{
		crawledUrl:make(map[string]bool),
	}
	Craw("https://golang.org/", fetch)
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
