package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/seventy-two/cara/web"

	"github.com/gorilla/mux"
)

const oops = "OOPSIE WOOPSIE!! Uwu We made a fucky wucky!! A wittle fucko boingo! The code monkeys at our headquarters are working VEWY HAWD to fix this!"

type httpHandler struct {
	eventSearchURL string
	eventDetailURL string
}

func newHTTPHandler(key, secret string) *httpHandler {
	m := md5.Sum([]byte("1" + secret + key))
	hM := hex.EncodeToString(m[:])
	search := fmt.Sprintf("https://gateway.marvel.com/v1/public/events%s?ts=1&apikey=%s&hash=%s%s", "%s", key, hM, "%s")
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
			writeMessage(w, http.StatusBadRequest, jsonMessage("missing event prefix"))
			return
		}
		es := &eventSearch{}
		u := fmt.Sprintf(h.eventSearchURL, "", "&nameStartsWith="+url.QueryEscape(prefix))
		err := web.GetJSON(u, es)
		if err != nil || es.Code != 200 || es.Data.Count < 1 {
			if err != nil {
				fmt.Println(err)
			}
			writeMessage(w, http.StatusBadRequest, jsonMessage(oops))
			return
		}
		c, err := populateComicsFromEvent(es.Data.Results[0].ID, h.eventDetailURL)
		if err != nil || len(c) == 0 {
			writeMessage(w, http.StatusBadRequest, jsonMessage(oops))
			return
		}
		eResponse := &EventResponse{
			Name:        es.Data.Results[0].Title,
			Description: es.Data.Results[0].Description,
			Comics:      c,
		}

		resp, err := json.Marshal(eResponse)
		if err != nil {
			writeMessage(w, http.StatusBadRequest, jsonMessage(oops))
			return
		}
		writeMessage(w, http.StatusOK, resp)
		return
	}
}

func getEventComicsByEventID(h *httpHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id, present := mux.Vars(r)["ID"]
		if !present {
			writeMessage(w, http.StatusBadRequest, jsonMessage("missing event ID"))
			return
		}
		es := &eventSearch{}
		u := fmt.Sprintf(h.eventSearchURL, "/"+id, "")
		err := web.GetJSON(u, es)
		if err != nil || es.Code != 200 || es.Data.Count < 1 {
			if err != nil {
				fmt.Println(err)
			}
			writeMessage(w, http.StatusBadRequest, jsonMessage(oops))
			return
		}
		c, err := populateComicsFromEvent(es.Data.Results[0].ID, h.eventDetailURL)
		if err != nil || len(c) == 0 {
			writeMessage(w, http.StatusBadRequest, jsonMessage(oops))
			return
		}
		eResponse := &EventResponse{
			Name:        es.Data.Results[0].Title,
			Description: es.Data.Results[0].Description,
			Comics:      c,
		}

		resp, err := json.Marshal(eResponse)
		if err != nil {
			writeMessage(w, http.StatusBadRequest, jsonMessage(oops))
			return
		}
		writeMessage(w, http.StatusOK, resp)
		return
	}
}

func getEvents(h *httpHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		es := &eventSearch{}
		u := fmt.Sprintf(h.eventSearchURL, "", "") + "&limit=35"
		err := web.GetJSON(u, es)
		if err != nil || es.Code != 200 || es.Data.Count < 1 {
			if err != nil {
				fmt.Println(err)
			}
			writeMessage(w, http.StatusBadRequest, jsonMessage(oops))
			return
		}

		for len(es.Data.Results) < es.Data.Total {
			esTemp := &eventSearch{}
			offset := strconv.Itoa(len(es.Data.Results))
			err = web.GetJSON(u+"&offset="+offset, esTemp)
			if err != nil || esTemp.Code != 200 || esTemp.Data.Count < 1 {
				if err != nil {
					fmt.Println(err)
				}
				writeMessage(w, http.StatusBadRequest, jsonMessage(oops))
				return
			}
			es.Data.Results = append(es.Data.Results, esTemp.Data.Results...)
		}
		var events []*Event
		for _, event := range es.Data.Results {
			events = append(events, &Event{
				Title: event.Title,
				ID:    event.ID,
			})
		}
		resp, err := json.Marshal(events)
		if err != nil {
			fmt.Println(err)
			writeMessage(w, http.StatusBadRequest, jsonMessage(oops))
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
