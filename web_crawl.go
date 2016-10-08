package main

import (
	"fmt"
	"time"
)

const crawl_num = 4

type State struct{
	url string
	done bool
}

type SafeTaskMap struct {
	m map[string]bool
	q chan string
}


type Fetcher interface {
	// Fetch 返回 URL 的 body 内容，并且将在这个页面上找到的 URL 放到一个 slice 中。
	Fetch(url string) (body string, urls []string, err error)
}

// Crawl 使用 fetcher 从某个 URL 开始递归的爬取页面，直到达到最大深度。
func Crawl(urlmap *SafeTaskMap, fetcher Fetcher, idle chan<- int, qstate chan<- *State, q_state_signal <-chan int, quit <-chan int) {
	// TODO: 并行的抓取 URL。
	// TODO: 不重复抓取页面。
    // 下面并没有实现上面两种情况：
	state_queue := make([]*State, 0, 10)
	idle_flag := false
	
	for {
		select {
		case <-quit:
			fmt.Println("goroutine exited!")
			return 
		case url := <-urlmap.q:
			if idle_flag{
				idle_flag = false
				idle <- -1
			}
			time.Sleep(2*time.Second)
			body, urls, err := fetcher.Fetch(url)
			state_queue = append(state_queue, &State{url, true})
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("found: %s %q\n", url, body)
			for _, u := range urls{
				state_queue = append(state_queue, &State{u, false})
			}
		case <-q_state_signal:
			if idle_flag{
				idle_flag = false
				idle <- -1
			}
			qstate <- state_queue[len(state_queue)]
			state_queue = state_queue[:len(state_queue) - 1]
		default:
			if len(state_queue) > 0{
				if idle_flag{
					idle_flag = false
					idle <- -1
				}
			}else if !idle_flag{
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
					//if idle_sum == crawl_num{
					//	fmt.Println("all crawl idle. --> quit")
					//	quit <- 0
					//	break
					//}
				case i := <-idle:
					idle_sum += i
			}
		}
	}()
	
	return idle
}

func StateProcessor(urlmap *SafeTaskMap) (chan <- *State, chan <- ){
	qstate := make(chan *State)
	go func(){
		for {
			
		}
		for state := range qstate{
			if state.done{
				urlmap.m[state.url] = true
			}else {
				done, ok := urlmap.m[state.url]
				if !ok || !done{
					urlmap.q <- state.url
				}
			}
			
		}
	}()
	
	return qstate
}

func main() {
	//urlmap := &SafeTaskMap{m:make(map[string]bool), q:make(chan string, 1)}
	urlmap := &SafeTaskMap{m:make(map[string]bool), q:make(chan string, crawl_num + 1)}
	quit := make(chan int)
	idle := IdleMonitor(time.Second, quit)
	urlmap.q <- "http://golang.org/"
	qstate := StateProcessor(urlmap)
	for i:=0; i < crawl_num; i++{
		go Crawl(urlmap, fetcher, idle, qstate, quit)
	}
	<-quit
}

// fakeFetcher 是返回若干结果的 Fetcher。
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

// fetcher 是填充后的 fakeFetcher。
var fetcher = fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
			"http://golang.org/doc/",
			"http://golang.org/about/",
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
