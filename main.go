package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/cloudflare/cfssl/log"
	"github.com/gorilla/mux"
	cli "github.com/jawher/mow.cli"
)

type appLink struct {
	name string
	url  string
}

var appMeta = struct {
	name        string
	description string
	discord     string
	maintainers string
	links       []appLink
}{
	name:        "Archivus",
	description: "A tool to help build comic book reading orders and lists",
	discord:     "https://discord.gg/F2cD4cN",
	maintainers: "github.com/seventy-two",
	links: []appLink{
		{name: "vcs", url: "https://github.com/seventy-two/archivus"},
	},
}

func main() {
	app := cli.App(appMeta.name, appMeta.description)
	publicKey := app.String(cli.StringOpt{
		Name:   "Marvel API Public Key",
		Value:  "",
		EnvVar: "MARVEL_PUBLIC_KEY",
	})
	privateKey := app.String(cli.StringOpt{
		Name:   "Marvel API Private Key",
		Value:  "",
		EnvVar: "MARVEL_PRIVATE_KEY",
	})
	addr := app.String(cli.StringOpt{
		Name:   "Archivus Address",
		Value:  "localhost:5000",
		EnvVar: "ARCHIVUS_ADDRESS",
	})

	app.Action = func() {
		archivus(app, *addr, *publicKey, *privateKey)
	}

	app.Run(os.Args)
}

func archivus(app *cli.Cli, addr, public, private string) {
	errors := make(chan error)

	router := mux.NewRouter()
	handler := newHTTPHandler(public, private)

	router.HandleFunc("/archivus/eventComicsByPrefix/{prefix}", getEventComicsByEventPrefix(handler)).Methods(http.MethodGet)
	router.HandleFunc("/archivus/eventComicsByEventID/{ID}", getEventComicsByEventID(handler)).Methods(http.MethodGet)
	router.HandleFunc("/archivus/getEvents", getEvents(handler)).Methods(http.MethodGet)

	server := &http.Server{Addr: addr, Handler: router}

	defer func() {
		log.Info("Stoping HTTP server")
		if err := server.Shutdown(context.Background()); err != nil {
			server.Close()
		}
	}()

	log.Infof("Starting HTTP server on %s", server.Addr)
	go func() {
		defer close(errors)
		errors <- server.ListenAndServe()
	}()

	// exit when appropriate
	exit := make(chan os.Signal)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-exit:
	}
}
