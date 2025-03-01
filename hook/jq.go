package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/fox-one/mixin-sdk-go/v2"
	"github.com/itchyny/gojq"
)

func (hdr *Handler) makeAppCard(pl any, r *http.Request) string {
	body := getString(pl, r.URL.Query().Get("body"))
	title := getString(pl, r.URL.Query().Get("title"))
	if title == "" && body == "" {
		return ""
	}
	link := getString(pl, r.URL.Query().Get("link"))
	link = url.QueryEscape(link)
	card := mixin.AppCardMessage{
		AppID:       hdr.mixin.ClientID,
		Title:       title,
		Description: body,
		Actions: mixin.AppButtonGroupMessage{{
			Label:  "View Details",
			Action: "https://mnm.sh/redirect?link=" + link,
			Color:  "#226F54",
		}},
	}
	b, _ := json.Marshal(card)
	return base64.RawURLEncoding.EncodeToString(b)
}

func getString(pl any, path string) string {
	if path == "" {
		return ""
	}
	query, err := gojq.Parse(path)
	if err != nil {
		return ""
	}
	it := query.Run(pl)
	v, found := it.Next()
	if !found {
		return ""
	}
	s, _ := v.(string)
	return s
}
