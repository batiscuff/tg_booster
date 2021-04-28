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
	"github.com/gammazero/workerpool"
)

func addView(proxy string, link string) {
	// Создание клиента и юзер-агента, добавление таймаута и прокси
	if utils.IsValidURL(link) != true {
		fmt.Println("Invalid post link! Example link: https://t.me/channel_name/1")
		os.Exit(1)
	}

	proxyUrl, err := url.Parse(proxy)
	ua := browser.Computer()
	client := &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
		Timeout:   30 * time.Second,
	}

	// Создание и конфигурация запроса
	link = link + "?embed=1"
	request, err := http.NewRequest("GET", link, nil)
	if err != nil {
		fmt.Println(err)
	}
	request.Header.Set("User-Agent", ua)

	var dataViewString string
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
	} else {
		// Вывод инфы в консоль
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("Error reading HTTP body. %q\n", err)
		}

		// Компиляция регулярки и её применение
		re := regexp.MustCompile(`data-view="(\w+)"`)

		if re.Match([]byte(body)) {
			dataViewString = string(re.FindSubmatch(body)[1])
		}
	}

	if response != nil && response.StatusCode == 200 {
		// Создание и конфигурация запроса
		request, err = http.NewRequest("GET", "https://t.me/v/?views="+dataViewString, nil)
		if err != nil {
			fmt.Println(err)
		}

		if len(response.Cookies()) != 0 {
			request.AddCookie(response.Cookies()[0])
		}
		request.Header.Set("X-Requested-With", "XMLHttpRequest")
		request.Header.Set("Referer", link)
		request.Header.Set("User-Agent", ua)

		response, err = client.Do(request)
		if err != nil {
			fmt.Println(err)
		} else {
			defer response.Body.Close()
			fmt.Printf("Response status code: %d\n", response.StatusCode)
		}

	}
}

func main() {
	fileName := flag.String("p", "", "Proxies file or link with http proxies ended .txt")
	workers := flag.Int("w", 50, "Count of workers on pool")
	postLink := flag.String("l", "", "Link on Telegram post for boost views. https://t.me/...")
	flag.Parse()

	proxies, err := utils.loadProxies(*fileName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// --- Workers Pool --
	wp := workerpool.New(*workers)
	for _, proxy := range proxies {
		proxy := proxy
		wp.Submit(func() {
			addView(proxy, *postLink)
		})
	}
	wp.StopWait()
}
