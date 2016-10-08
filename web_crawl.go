package main

import (
    "container/list"
	"fmt"
	"time"
)

const crawl_num = 5

type SignalUrlChan struct{
    url chan string
    signal chan int
}

type TaskMap struct {
	m map[string]bool
	q chan string
}


type Fetcher interface {
	Fetch(url string) (body string, urls []string, err error)
}

func Crawl(urlmap *TaskMap, fetcher Fetcher, idle chan<- int, complete *SignalUrlChan, todo *SignalUrlChan) {
	complete_list := list.New()
    todo_list := list.New()
	idle_flag := false
	
	for {
        if complete_list.Len() > 0{
            <-complete.signal
            e := complete_list.Front()
            complete.url <- e.Value.(string)
            complete_list.Remove(e)
            continue
        }else if todo_list.Len() > 0{
            <- todo.signal
            e := todo_list.Front()
            todo.url <- e.Value.(string)
            todo_list.Remove(e)
        }
 
		select {
		case url := <-urlmap.q:
			if idle_flag{
				idle_flag = false
				idle <- -1
			}
			body, urls, err := fetcher.Fetch(url)
            complete_list.PushBack(url)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("found: %s %q\n", url, body)
			for _, u := range urls{
                todo_list.PushBack(u)
			}
		default:
			if !idle_flag{
				idle_flag = true
				idle <- 1
			}
		}
	}
}

func IdleMonitor(interval time.Duration, quit chan<- int) chan<- int{
	//idle := make(chan int)
	idle := make(chan int, crawl_num)
	idle_sum := 0
	ticker := time.NewTicker(interval)
	
	go func(){
		for{
			select{
				case <-ticker.C:
					fmt.Printf("[IDLE]->%v\n", idle_sum)
					if idle_sum == crawl_num{
						fmt.Println("all crawl idle. --> quit")
						quit <- 0
						break
					}
				case i := <-idle:
					idle_sum += i
			}
		}
	}()
	
	return idle
}

func StateProcessor(urlmap *TaskMap) (*SignalUrlChan, *SignalUrlChan){
	complete := &SignalUrlChan{url:make(chan string, crawl_num), signal:make(chan int, crawl_num)}
    todo := &SignalUrlChan{url:make(chan string, crawl_num), signal:make(chan int, crawl_num)}
    go func(){
		for {
            select{
                case url := <-complete.url:
                    urlmap.m[url] = true
                case url := <-todo.url:
                    _, ok := urlmap.m[url]
                    if !ok {
                        //fmt.Println("[PUSH]", url, ok)
                        urlmap.q <- url
                    }
                case complete.signal <- 0:
                case todo.signal <- 0:
            }
		}
	}()
	
	return complete, todo
}

func main() {
	//urlmap := &TaskMap{m:make(map[string]bool), q:make(chan string, 1)}
	urlmap := &TaskMap{m:make(map[string]bool), q:make(chan string, crawl_num + 1)}
	quit := make(chan int)
	idle := IdleMonitor(time.Second, quit)
	urlmap.q <- "http://golang.org/"
	complete, todo := StateProcessor(urlmap)
	for i:=0; i < crawl_num; i++{
		go Crawl(urlmap, fetcher, idle, complete, todo)
	}
	<-quit
    for k, _ := range urlmap.m{
        fmt.Println(k)
    }
}

// fakeFetcher 是返回若干结果的 Fetcher。
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
    time.Sleep(time.Second)
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher 是填充后的 fakeFetcher。
var fetcher = fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
			"http://golang.org/doc/",
			"http://golang.org/about/",
            "http://golang.org/abc/",
            "http://golang.org/hello/",
            "http://golang.org/test/",
            "http://golang.org/proxy/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
			"http://golang.org/pkg/time/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/time/": &fakeResult{
		"Package time",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
			"http://www.baidu.com/",
		},
	},
}
