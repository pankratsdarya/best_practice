package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/html"
)

// парсим страницу
func parse(ctxt context.Context, url string) (*html.Node, error) {
	select {
	case <-ctxt.Done():
		return nil, fmt.Errorf("time's up")

	default:
		// что здесь должно быть вместо http.Get? :)
		var netClient = &http.Client{
			Timeout: time.Second * 4,
		}
		r, err := netClient.Get(url)

		if err != nil {
			return nil, fmt.Errorf("can't get page")
		}
		b, err := html.Parse(r.Body)
		if err != nil {
			return nil, fmt.Errorf("can't parse page")
		}
		return b, err
	}
}

// ищем заголовок на странице
func pageTitle(ctxt context.Context, n *html.Node) string {
	select {
	case <-ctxt.Done():
		return ""

	default:
		var title string
		if n.Type == html.ElementNode && n.Data == "title" {
			return n.FirstChild.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			title = pageTitle(ctxt, c)
			if title != "" {
				break
			}
		}
		return title
	}
}

// ищем все ссылки на страницы. Используем мапку чтобы избежать дубликатов
func pageLinks(ctxt context.Context, links map[string]struct{}, n *html.Node) map[string]struct{} {
	select {
	case <-ctxt.Done():
		return nil
	default:

		if links == nil {
			links = make(map[string]struct{})
		}

		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key != "href" {
					continue
				}

				// костылик для простоты
				if _, ok := links[a.Val]; !ok && len(a.Val) > 2 && a.Val[:2] == "//" {
					links["http://"+a.Val[2:]] = struct{}{}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			links = pageLinks(ctxt, links, c)
		}
		return links
	}
}
