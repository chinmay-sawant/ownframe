package ownframe

import "github.com/chinmay-sawant/ownframe/internal/page"

func wrapPage(p *page.Page, err error) (*Page, error) {
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, nil
	}

	return &Page{Page: p}, nil
}

func innerPage(p *Page) *page.Page {
	if p == nil {
		return nil
	}

	return p.Page
}
