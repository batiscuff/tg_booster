package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	utils "github.com/batiscuff/tg_booster/boosterutils"
	
    browser "github.com/EDDYCJY/fake-useragent"
    . "github.com/logrusorgru/aurora"
	"github.com/gammazero/workerpool"
)

var goodProxies []string

func addView(proxy string, link string) {
	// Checking the link to the post from the telegram channel
	if utils.CheckPostLink(link) != true {
		fmt.Println(Bold(Red("Invalid post link! Example link: https://t.me/channel_name/1")))
		os.Exit(1)
	}
    link = link + "?embed=1"

    // Creating a client and User-Agent, adding a timeout and proxy
	proxyUrl, _ := url.Parse(proxy)
	ua := browser.Computer()
	client := &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
		Timeout:   30 * time.Second,
	}

	// Creating and configure 1 request
	request, _ := http.NewRequest("GET", link, nil)
	request.Header.Set("User-Agent", ua)

	// Sending 1 request and response processing
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(Sprintf(Red(err)))
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
        fmt.Println(Sprintf("Error reading HTTP body. %q", Red(err)))
		return
	}

    // Compiling and using regex
	var dataViewString string
	re := regexp.MustCompile(`data-view="(\w+)"`)
	if re.Match([]byte(body)) {
		dataViewString = string(re.FindSubmatch(body)[1])
	} else {
		fmt.Println("Data-views not found.")
		return
	}

	// Configure 2 request
	request, _ = http.NewRequest("GET", "https://t.me/v/?views="+dataViewString, nil)
	if len(response.Cookies()) != 0 {
		request.AddCookie(response.Cookies()[0])
	}
    request.Header.Set("User-Agent", ua)
	request.Header.Add("X-Requested-With", "XMLHttpRequest")
	request.Header.Add("Referer", link)
    
    // Sending 2 request
	response, err = client.Do(request)
	if err != nil {
		fmt.Println(Red(err))
		return
	}
	defer response.Body.Close()
    
    if response.StatusCode == 200 {
    	fmt.Println(Sprintf(Yellow("Views added! [%s]"), Green(proxy)))
    	goodProxies = append(goodProxies, proxy)
    }
}

func main() {
	fileName := flag.String("p", "", "Proxies file or link with http proxies ended .txt")
	workers := flag.Int("w", 50, "Count of workers on pool")
	postLink := flag.String("l", "", "Link on Telegram post for boost views. https://t.me/...")
	flag.Parse()

	proxies, err := utils.LoadProxies(*fileName)
	if err != nil {
		fmt.Println(err)
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

    var resultString = "Proxies count: %d\tViews count: %d"
    var lenProxies, lenGoodProxies = len(proxies), len(goodProxies)
	fmt.Println(Sprintf(Bold(Magenta(resultString)), Cyan(lenProxies), Cyan(lenGoodProxies)))
	fmt.Println(Sprintf(Bold(Magenta("Run time: %s")), Cyan(elapsed)))
}
