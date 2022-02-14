package pusher

import (
	"net/http"
	"net/url"
)

type Item interface {
	GetBody() []byte
	GetMethod() string
	GetHeader() http.Header
	GetURL() *url.URL
	GetEventID() string
}

type item struct {
	body      []byte
	method    string
	header    http.Header
	originURL *url.URL
	eventID   string
}

func (i *item) GetBody() []byte {
	return i.body
}

func (i *item) GetMethod() string {
	return i.method
}

func (i *item) GetHeader() http.Header {
	return i.header
}

func (i *item) GetURL() *url.URL {
	return i.originURL
}

func (i *item) GetEventID() string {
	return i.eventID
}

func NewItem(
	body []byte,
	method string,
	header http.Header,
	originURL *url.URL,
	eventID string,
) Item {
	return &item{
		body:      body,
		method:    method,
		eventID:   eventID,
		header:    header,
		originURL: originURL,
	}
}
