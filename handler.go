package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/seventy-two/cara/web"

	"github.com/gorilla/mux"
)

type httpHandler struct {
	eventSearchURL string
	eventDetailURL string
}

func newHTTPHandler(key, secret string) *httpHandler {
	m := md5.Sum([]byte("1" + secret + key))
	hM := hex.EncodeToString(m[:])
	search := fmt.Sprintf("https://gateway.marvel.com/v1/public/events?ts=1&apikey=%s&hash=%s&nameStartsWith=%s", key, hM, "%s")
	event := fmt.Sprintf("https://gateway.marvel.com/v1/public/events/%s/comics?ts=1&apikey=%s&hash=%s&limit=100%s", "%d", key, hM, "%s")
	return &httpHandler{
		eventSearchURL: search,
		eventDetailURL: event,
	}
}

func getEventComicsByEventPrefix(h *httpHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		prefix, present := mux.Vars(r)["prefix"]
		if !present {
			writeMessage(w, http.StatusBadRequest, jsonMessage("missing eventprefix"))
			return
		}
		es := &eventSearch{}
		u := fmt.Sprintf(h.eventSearchURL, url.QueryEscape(prefix))
		err := web.GetJSON(u, es)
		if err != nil || es.Code != 200 || es.Data.Count < 1 {
			if err != nil {
				fmt.Println(err)
			}
			writeMessage(w, http.StatusBadRequest, jsonMessage("OOPSIE WOOPSIE!! Uwu We made a fucky wucky!! A wittle fucko boingo! The code monkeys at our headquarters are working VEWY HAWD to fix this!"))
			return
		}
		c, err := populateComicsFromEvent(es.Data.Results[0].ID, h.eventDetailURL)
		if err != nil || len(c) == 0 {
			writeMessage(w, http.StatusBadRequest, jsonMessage("OOPSIE WOOPSIE!! Uwu We made a fucky wucky!! A wittle fucko boingo! The code monkeys at our headquarters are working VEWY HAWD to fix this!"))
			return
		}
		eResponse := &EventResponse{
			Name:        es.Data.Results[0].Title,
			Description: es.Data.Results[0].Description,
			Comics:      c,
		}

		resp, err := json.Marshal(eResponse)
		if err != nil {
			writeMessage(w, http.StatusBadRequest, jsonMessage("OOPSIE WOOPSIE!! Uwu We made a fucky wucky!! A wittle fucko boingo! The code monkeys at our headquarters are working VEWY HAWD to fix this!"))
			return
		}
		writeMessage(w, http.StatusOK, resp)
		return
	}
}

func jsonMessage(message string) []byte {
	format := "{\"message\":\"%s\"}"
	return []byte(fmt.Sprintf(format, message))
}

func writeMessage(w http.ResponseWriter, code int, message []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(message)
}
