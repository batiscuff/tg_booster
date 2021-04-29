package boosterutils

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
    "regexp"
)

func IsValidURL(inputUrl string) bool {
	_, err := url.ParseRequestURI(inputUrl)
	if err != nil {
		return false
	}
	outputUrl, err := url.Parse(inputUrl)
	if err != nil || outputUrl.Scheme == "" || outputUrl.Host == "" {
		return false
	}
	return true
}

func CheckPostLink(link string) bool {
    re := regexp.MustCompile(`https://t\.me/\w+/\d+`)
    if re.MatchString(link) {
        return true
    }
    return false
}

func LoadProxies(fileName string) (proxies []string, err error) {
	var body io.Reader
	// Загрузка прокси из линка и из файла
	if IsValidURL(fileName) {
		resp, err := http.Get(fileName)
		if err != nil {
			return nil, fmt.Errorf("Proxies from link not loaded! %q", err)
		}
		defer resp.Body.Close()
		body = resp.Body
	} else {
		fileObj, err := os.Open(fileName)
		if err != nil {
			return nil, fmt.Errorf("Proxies from file not loaded! %q", err)
		}
		defer fileObj.Close()
		body = fileObj
	}

	var lines []string
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) != 0 { // index out of range [0] with length 0
			if IsValidURL(line) {
				lines = append(lines, line)
			} else {
				line = "http://" + line
				lines = append(lines, line)
			}
		}
	}
	return lines, scanner.Err()
}
