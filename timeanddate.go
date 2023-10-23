package timeanddate

import (
	"gopkg.in/resty.v1"
)

var DefaultClient = New()

func New() *Client {
	r := resty.New()

	// Get(path string) redirets
	r.SetRedirectPolicy(resty.FlexibleRedirectPolicy(2))
	r.SetHeader("Accept-Language", "en")

	return &Client{
		resty: r,
	}
}

type Client struct {
	resty *resty.Client
}
