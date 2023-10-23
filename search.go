package timeanddate

import (
	// "gopkg.in/resty.v1"

	"bufio"
	"bytes"
	"errors"
	"net/url"
	"strings"
)

var (
	ErrInvalidResponse = errors.New("Received invalid response")
)

type SearchResult struct {
	Path        string // e.g. /worldclock/@2830841
	Unused1     string // e.g. 5
	CountryCode string // e.g. de
	Unused2     string // e.g. b
	City        string // e.g. Bezirk Spandau (fifth-order administrative division)

	State       string // e.g. Berlin
	Country     string // e.g. Germany
	CountryFlag string // e.g. //c.tadst.com/gfx/n/fl/16/de.png
	Unused3     string // e.g. ""
	Unused4     string // e.g. ""
	Unused5     string // e.g. "p"
	Unused6     string // e.g. ""
}

// performs search with DefaultClient
func Search(query string) (r []SearchResult, err error) {
	return DefaultClient.Search(query)
}

// https://www.timeanddate.com/scripts/completion.php?query=!!QUERY¡¡&xd=1&mode=ci
func (c *Client) Search(query string) (r []SearchResult, err error) {
	v := url.Values(map[string][]string{
		"xd":    {"1"},
		"query": {query},
		"mode":  {"ci"},
	})

	url := "https://www.timeanddate.com/scripts/completion.php?" + v.Encode()
	res, err := c.resty.R().Get(url)
	if err != nil {
		return
	}

	s := bufio.NewScanner(bytes.NewReader(res.Body()))
	for s.Scan() {
		row := strings.Split(s.Text(), string([]byte{0x09}))

		// ending row
		if len(row) == 1 {
			continue
		}

		if len(row) != 12 {
			return nil, ErrInvalidResponse
		}

		// example split row
		// {"/worldclock/@2830841", "5", "de", "b",
		//  "Bezirk Spandau (fifth-order administrative division)",
		//  "Berlin", "Germany", "//c.tadst.com/gfx/n/fl/16/de.png",
		//  "", "", "p", ""}
		r = append(r, SearchResult{ // wow much code, so quality
			Path:        row[0],
			Unused1:     row[1],
			CountryCode: row[2],
			Unused2:     row[3],
			City:        row[4],
			State:       row[5],
			Country:     row[6],
			CountryFlag: row[7],
			Unused3:     row[8],
			Unused4:     row[9],
			Unused5:     row[10],
			Unused6:     row[11],
		})
	}

	return
}
