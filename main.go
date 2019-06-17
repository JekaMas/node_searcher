package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"node_searcher/searcher"

	"github.com/valyala/fasthttp"
)

func main() {
	jobs := make(chan string, searcher.ParseJobs)
	wg := &sync.WaitGroup{}
	for i := 0; i < searcher.GetHashers; i++ {
		wg.Add(1)
		go Search(jobs, wg)
	}

	wgParsers := &sync.WaitGroup{}
	for i := 0; i < searcher.GetParsers; i++ {
		wgParsers.Add(1)
		go Parse(jobs, wgParsers)
	}

	wg.Wait()
	close(jobs)

	wgParsers.Wait()

	ioutil.WriteFile("parsed_nodes.txt", searcher.PrintNodesMap(), 0644)
}

var shouldWait = new(int32)
var doneFoundCount = new(int32)
var doneCount = new(int32)

func Parse(ch <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for id := range ch {
		if atomic.LoadInt32(shouldWait) == 1 {
			time.Sleep(time.Second)
			atomic.StoreInt32(shouldWait, 0)
		}

		url := searcher.NodeUrl + id
		res := doRequest(url)

		searcher.NodesMapLock.RLock()
		node := searcher.NodesMap[id]
		searcher.NodesMapLock.RUnlock()

		// e.g. [eth/63 eth/62]
		capabilitiesStart := `<th scope="row" class="text-right">Capabilities</th><td>`
		idxCapabilitiesStart := bytes.Index(res, []byte(capabilitiesStart)) + len(capabilitiesStart)
		if idxCapabilitiesStart < 0 {
			fmt.Println("Cant find Capabilities(1):", id, string(res))
			atomic.StoreInt32(shouldWait, 1)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		capabilitiesEnd := "</td>"
		if len(res) <= idxCapabilitiesStart {
			fmt.Println("Cant find Capabilities(2):", id, string(res))
			atomic.StoreInt32(shouldWait, 1)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		idxCapabilitiesEnd := bytes.Index(res[idxCapabilitiesStart:], []byte(capabilitiesEnd)) + idxCapabilitiesStart
		if len(res) <= idxCapabilitiesEnd {
			fmt.Println("Cant find Capabilities(3):", id, string(res))
			atomic.StoreInt32(shouldWait, 1)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		node.Capabilities = string(res[idxCapabilitiesStart:idxCapabilitiesEnd])

		// filter stage
		if strings.Index(node.Capabilities, "les/2") >= 0 ||
			strings.Index(node.Capabilities, "les:2") >= 0 ||
			strings.Index(node.Capabilities, "les2") >= 0 {

			searcher.NodesByClientLock.Lock()

			key := node.Client
			if len(node.Client) == 0 {
				key = node.ClientVersion
			}
			searcher.NodesByClient[key] = append(searcher.NodesByClient[key], node)
			searcher.NodesByClientLock.Unlock()

			doneFoundCountCurrent := atomic.AddInt32(doneFoundCount, 1)
			if doneFoundCountCurrent%10 == 0 {
				fmt.Println("Storing", doneFoundCountCurrent)
				ioutil.WriteFile("parsed_nodes.txt", searcher.PrintNodesMap(), 0644)
			}
		}

		doneFoundCountCurrent := atomic.AddInt32(doneCount, 1)
		if doneFoundCountCurrent%1000 == 0 {
			fmt.Println("\tDone parsing:", doneFoundCountCurrent)
		}
	}
}

func Search(ch chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		to, err := searcher.GetJobNumber()
		if err != nil {
			fmt.Println(err)
			return
		}

		if to%1000 == 0 {
			fmt.Println("Done hashes:", to)
		}

		url := searcher.ServiceUrl + strconv.Itoa(to)
		res := doRequest(url)

		nodes := &searcher.ReqNodes{}
		nodes.Decode(res)

		for i := range nodes.Data {
			searcher.NodesMapLock.Lock()
			searcher.NodesMap[nodes.Data[i].ID] = nodes.Data[i]
			searcher.NodesMapLock.Unlock()

			ch <- string(nodes.Data[i].ID)
		}
	}

	fmt.Println("\tParser have done its jobs")
}

func doRequest(url string) []byte {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)

	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	client.Do(req, resp)

	return resp.Body()
}
