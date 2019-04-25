package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/seventy-two/cara/web"
)

func populateComicsFromEvent(id int, detailURL string) ([]*Comic, error) {

	ed := &eventDetail{}
	url := fmt.Sprintf(detailURL, id, "")
	err := web.GetJSON(url, ed)
	if err != nil {
		return nil, err
	}

	if ed.Data.Count < ed.Data.Total {
		ed2 := &eventDetail{}
		err = web.GetJSON(fmt.Sprintf(detailURL, id, "&offset=100"), ed2)
		if err != nil {
			return nil, err
		}
		ed.Data.Results = append(ed.Data.Results, ed2.Data.Results...)
	}

	var comix []*Comic
	for _, c := range ed.Data.Results {
		if strings.ToLower(c.Format) != "comic" {
			continue
		}
		var d time.Time
		for _, relDate := range c.Dates {
			if relDate.Type == "onsaleDate" {
				d, err = time.Parse("2006-01-02T15:04:05-0700", relDate.Date)
				if err != nil {
					return nil, err
				}
			}
		}
		var cURL string
		for _, u := range c.Urls {
			if u.Type == "detail" {
				cURL = u.URL
			}
		}

		comix = append(comix, &Comic{
			Title: c.Title,
			URL:   cURL,
			Image: c.Thumbnail.Path + "." + c.Thumbnail.Extension,
			Date:  d,
		})
	}
	sort.Slice(comix, func(i, j int) bool { return comix[i].Date.Before(comix[j].Date) })
	return comix, nil
}
