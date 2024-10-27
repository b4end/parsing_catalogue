package main

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
)

type Product struct {
	Name        string
	Description string
}

func main() {
	product_data := parsing_catalogue()

	for i := 0; i < len(product_data); i++ {
		fmt.Printf("id %d:\nНазвание: %s\nОписание: %s\n\n", i, product_data[i].Name, product_data[i].Description)
	}
}

func get_html(url string) (*html.Node, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка при получении страницы: %s", resp.Status)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func parsing_catalogue() []Product {
	url := "https://www.incotexcom.ru/catalogue"
	var links []string

	doc, err := get_html(url)
	if err != nil {
		fmt.Println(err)
	}

	var get_catalogues func(*html.Node)
	get_catalogues = func(h *html.Node) {
		if h.Type == html.ElementNode && h.Data == "a" {
			for first_h := h.FirstChild; first_h != nil; first_h = first_h.NextSibling {
				if first_h.Type == html.ElementNode && first_h.Data == "img" {
					for _, at := range first_h.Attr {
						if at.Key == "alt" && at.Val != "Программное обеспечение" {
							links = append(links, h.Attr[0].Val)
						}
					}
				}
			}
		}

		for first_h := h.FirstChild; first_h != nil; first_h = first_h.NextSibling {
			get_catalogues(first_h)
		}
	}

	get_catalogues(doc)

	return parsing_catalogues(links)
}

func parsing_catalogues(urls []string) []Product {
	var links []string

	for _, url := range urls {
		doc, err := get_html(url)
		if err != nil {
			fmt.Println("Ошибка при получении страницы:", err)
			continue
		}

		var get_links func(*html.Node)
		get_links = func(h *html.Node) {
			if h.Type == html.ElementNode && h.Data == "header" {
				for _, atr := range h.Attr {
					if atr.Key == "class" && atr.Val == "product-intro__slider" {
						for first_h := h.FirstChild; first_h != nil; first_h = first_h.NextSibling {
							if first_h.Type == html.ElementNode && first_h.Data == "a" {
								for _, linkAttr := range first_h.Attr {
									if linkAttr.Key == "href" {
										links = append(links, "https://www.incotexcom.ru"+linkAttr.Val)
									}
								}
							}
						}
					}
				}
			}

			for first_h := h.FirstChild; first_h != nil; first_h = first_h.NextSibling {
				get_links(first_h)
			}
		}

		get_links(doc)
	}

	return parsing_page(links)
}

func parsing_page(links []string) []Product {
	var products []Product

	for _, url := range links {
		doc, err := get_html(url)
		if err != nil {
			fmt.Println("Ошибка при получении страницы:", err)
			continue
		}

		var title, text string

		var get_product func(*html.Node)
		get_product = func(h *html.Node) {
			if h.Type == html.ElementNode {
				switch h.Data {
				case "h1":
					if h.FirstChild != nil && h.FirstChild.Type == html.TextNode {
						title = h.FirstChild.Data
					}
				case "div":
					for _, atr := range h.Attr {
						if atr.Key == "class" && atr.Val == "col-xs-12 col-sm-6" {
							for first_h := h.FirstChild; first_h != nil; first_h = first_h.NextSibling {
								if first_h.Type == html.ElementNode && first_h.Data == "p" {
									var extract_description func(*html.Node)
									extract_description = func(p *html.Node) {
										if p.Type == html.TextNode {
											text += p.Data
										}

										for child := p.FirstChild; child != nil; child = child.NextSibling {
											extract_description(child)
										}
									}

									text += " "
									extract_description(first_h)
								}
							}
						}
					}
				}
			}

			for first_h := h.FirstChild; first_h != nil; first_h = first_h.NextSibling {
				get_product(first_h)
			}
		}

		get_product(doc)

		text_start := 0
		for text_start < len(text) && text[text_start] == ' ' {
			text_start++
		}

		p := Product{title, text[text_start:]}
		products = append(products, p)
	}

	return products
}
