package client

import (
	"net/url"
	"strconv"
)

// By default, the number of objects returned per page is 50.
// It can be adjusted by adding the 'count' parameter in the query string.
// The maximum number of objects that can be retrieved per page vary between 50 and 100 depending on the endpoint.
// https://developers.pandadoc.com/reference/about - Look for the specific endpoint for more details.

const ItemsPerPage = 50 // Common default number of items per page as per PandaDoc API

// PageOptions is options for list method of paginatable resources.
// It's used to create query string.
type PageOptions struct {
	Count int `url:"count,omitempty"`
	Page  int `url:"page,omitempty"`
}

type ReqOpt func(reqURL *url.URL)

// count : items per page.
func WithPageLimit(count int) ReqOpt {
	if count <= 0 || count > ItemsPerPage {
		count = ItemsPerPage
	}
	return WithQueryParam("count", strconv.Itoa(count))
}

// page: Number for the page (inclusive). The page number starts with 1.
// If page is 0, first page is assumed.
func WithPage(page int) ReqOpt {
	if page == 0 {
		page = 1
	}
	return WithQueryParam("page", strconv.Itoa(page))
}

func WithQueryParam(key string, value string) ReqOpt {
	return func(reqURL *url.URL) {
		q := reqURL.Query()
		q.Set(key, value)
		reqURL.RawQuery = q.Encode()
	}
}
