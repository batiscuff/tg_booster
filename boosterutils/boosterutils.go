package boosterutils

import "net/url"

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
