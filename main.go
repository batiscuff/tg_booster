package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	utils "github.com/batiscuff/tg_booster/boosterutils"

	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/gammazero/workerpool"
	. "github.com/logrusorgru/aurora"
)

var (
	goodProxies []string
	re          *regexp.Regexp = regexp.MustCompile(`data-view="(\w+)"`)
)

func addView(proxy string, link string) {
	// Creating User-Agent
	randomUA := browser.NewBrowser(browser.Client{}, browser.Cache{}).Random()

	// Creating a client and adding a timeout and proxy
	proxyUrl, _ := url.Parse(proxy)
	client := &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
	}

	// Creating and configure 1 request
	request, _ := http.NewRequest("GET", link+"?embed=1", nil)
	request.Close = true
	request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/jxl,image/webp,*/*;q=0.8")
	request.Header.Add("Accept-Encoding", "gzip, deflate, br")
	request.Header.Add("User-Agent", randomUA)

	// Sending 1 request and response processing
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(Sprintf(Red(err)))
		return
	}
	defer response.Body.Close()

	// Check that the server actually sent compressed data
	var reader io.ReadCloser
	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(response.Body)
		defer reader.Close()
	default:
		reader = response.Body
	}

	// Read resp body
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		fmt.Println(Sprintf("Error reading HTTP body. %q", Red(err)))
		return
	}

	// Finding data-view by regexp
	var dataViewString string
	if re.Match([]byte(data)) {
		dataViewString = string(re.FindSubmatch(data)[1])
	} else {
		fmt.Println("Data-views not found.")
		return
	}

	// Configure 2 request
	request, _ = http.NewRequest("GET", "https://t.me/v/?views="+dataViewString, nil)
	request.Close = true
	if len(response.Cookies()) != 0 {
		request.AddCookie(response.Cookies()[0])
	}
	request.Header.Add("Accept", "*/*")
	request.Header.Add("Accept-Encoding", "gzip, deflate, br")
	request.Header.Add("Referer", link)
	request.Header.Add("User-Agent", randomUA)
	request.Header.Add("X-Requested-With", "XMLHttpRequest")

	// Sending 2 request
	response, err = client.Do(request)
	if err != nil {
		fmt.Println(Red(err))
		return
	}
	defer response.Body.Close()

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		fmt.Println(Sprintf(Yellow("Views added! [%s]"), Green(proxy)))
		goodProxies = append(goodProxies, proxy)
	}
}

func main() {
	fileName := flag.String("p", "", "Proxies file or link with http proxies ended .txt")
	workers := flag.Int("w", 50, "Count of workers on pool")
	postLink := flag.String("l", "", "Link on Telegram post for boost views. https://t.me/...")
	flag.Parse()

	// Check the proxies file/link
	proxies, err := utils.LoadProxies(*fileName)
	if err != nil {
		fmt.Println(Bold(Red(err)))
		os.Exit(1)
	}

	// Checking the link to the post from the telegram channel
	if utils.CheckPostLink(*postLink) != true {
		fmt.Println(Bold(Red("Invalid post link! Example link: https://t.me/channel_name/1")))
		os.Exit(1)
	}

	start := time.Now()
	// --- Workers Pool --
	wp := workerpool.New(*workers)
	for _, proxy := range proxies {
		proxy := proxy
		wp.Submit(func() {
			addView(proxy, *postLink)
		})
	}
	wp.StopWait()
	elapsed := time.Since(start)

	var lenProxies, lenGoodProxies = len(proxies), len(goodProxies)
	fmt.Println(Sprintf(Bold(Magenta("Proxies count: %d\tViews count: %d")), Cyan(lenProxies), Cyan(lenGoodProxies)))
	fmt.Println(Sprintf(Bold(Magenta("Run time: %s")), Cyan(elapsed)))
}
