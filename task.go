package main

import (
	"context"
	"log"
)

func Get(ctx context.Context, url string) (body string, err error)

func ParseHTML(body string) (urls []string)

func parseHost(url string) (host string)

func prepareURLMap(ctx context.Context, baseURL string) map[string][]string {
	var urls []string
	var urlMap = make(map[string][]string)

	baseHost := parseHost(baseURL)

	urls = append(urls, baseURL)

	for len(urls) > 0 {
		current := urls[0]
		urls = urls[1:]

		if _, seen := urlMap[current]; seen {
			continue
		}

		body, err := Get(ctx, current)
		if err != nil {
			log.Println(err.Error())
			urlMap[current] = []string{}
			continue
		}

		foundUrls := ParseHTML(body)

		var sameHostUrls []string

		for _, url := range foundUrls {
			if parseHost(url) == baseHost {
				sameHostUrls = append(sameHostUrls, url)
			}
		}
		urlMap[current] = sameHostUrls

		for _, url := range sameHostUrls {
			if _, seen := urlMap[url]; !seen {
				urls = append(urls, url)
			}
		}
	}
	return urlMap
}
