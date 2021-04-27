package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"

	utils "github.com/batiscuff/tg_booster/boosterutils"

	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/gammazero/workerpool"
)

func addView(proxy string) {
	// Создание клиента и юзер-агента, добавление таймаута и прокси
	proxyUrl, err := url.Parse(proxy)
	ua := browser.Computer()
	client := &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
		Timeout:   30 * time.Second,
	}

	// Создание и конфигурация запроса
	requestUrl := "https://t.me/parsing_conf/108?embed=1"
	request, err := http.NewRequest("GET", requestUrl, nil)
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
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("Error reading HTTP body. %q\n", err)
		}

		// Компиляция регулярки и её применение
		re := regexp.MustCompile(`data-view="(\w+)"`)

		if re.Match([]byte(body)) {
			dataViewString = string(re.FindSubmatch(body)[1])
		}
		response.Body.Close()
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
		request.Header.Set("Referer", requestUrl)
		request.Header.Set("User-Agent", ua)

		response, err = client.Do(request)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("Response status code: %d\n", response.StatusCode)
			response.Body.Close()
		}

	}
}


func loadProxies() (proxies []string, err error) {
	// var fileObj os.Stdin
	var body io.Reader
	if len(os.Args) <= 1 || (os.Args[1] == "-h" || os.Args[1] == "--help") {
		err = fmt.Errorf("Запуск приложения возможен с названием файла или ссылкой на прокси!")
		return nil, err
	} else if len(os.Args) >= 2 {
		filename := os.Args[1]
		// Загрузка прокси из линка и из файла
		if utils.IsValidURL(filename) {
			resp, err := http.Get(filename)
			if err != nil {
				return nil, fmt.Errorf("Прокси из ссылки небыли загружены!")
			}
			defer resp.Body.Close()
			body = resp.Body
		} else {
			fileObj, err := os.Open(filename)
			if err != nil {
				log.Fatal(err)
			}
			defer fileObj.Close()
			body = fileObj
		}
	}

	var lines []string
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) != 0 { // index out of range [0] with length 0
			if utils.IsValidURL(line) {
				lines = append(lines, line)
			} else {
				line = "http://" + line
				lines = append(lines, line)
			}
		}
	}
	return lines, scanner.Err()
}

func main() {
	proxies, err := loadProxies()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// --- Workers Pool --
	var workers int
	if len(os.Args) == 3 {
		workers, err = strconv.Atoi(os.Args[2])
	} else {
		workers = 100
	}

	wp := workerpool.New(workers)
	for _, proxy := range proxies {
		proxy := proxy
		wp.Submit(func() {
			addView(proxy)
		})
	}
	wp.StopWait()
}
