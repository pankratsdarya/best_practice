package main

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/net/html"
)

// парсим страницу
func parse(url string, ctxt context.Context) (*html.Node, error) {
	select {
	case <-ctxt.Done():
		return nil, fmt.Errorf("time's up")

	default:
		// что здесь должно быть вместо http.Get? :)
		r, err := http.Get(url)
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
func pageTitle(n *html.Node, ctxt context.Context) string {
	select {
	case <-ctxt.Done():
		return ""

	default:
		var title string
		if n.Type == html.ElementNode && n.Data == "title" {
			return n.FirstChild.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			title = pageTitle(c, ctxt)
			if title != "" {
				break
			}
		}
		return title
	}
}

// ищем все ссылки на страницы. Используем мапку чтобы избежать дубликатов
func pageLinks(links map[string]struct{}, n *html.Node, ctxt context.Context) map[string]struct{} {
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
			links = pageLinks(links, c, ctxt)
		}
		return links
	}
}
