package main

import (
	"context"
	"flag"
	"github.com/gissleh/sarfya"
	"github.com/gissleh/sarfya-service/adapters/fwewdictionary"
	"github.com/gissleh/sarfya-service/adapters/sourcestorage"
	"github.com/gissleh/sarfya-service/adapters/templfrontend"
	"github.com/gissleh/sarfya-service/adapters/webapi"
	"github.com/gissleh/sarfya/adapters/placeholderdictionary"
	"github.com/gissleh/sarfya/sarfyaservice"
	"log"
	"strings"
)

var flagSourceDir = flag.String("source-dir", "./data", "Source directory")
var flagListenAddr = flag.String("listen", ":8080", "Listen address")
var flagLitxapApi = flag.String("litxap-api", "http://localhost:8081", "Litxap root address")

func main() {
	dict := sarfya.CombinedDictionary{
		sarfya.WithDerivedPoS(fwewdictionary.Global()),
		placeholderdictionary.New(),
	}

	storage, err := sourcestorage.Open(context.Background(), *flagLitxapApi, *flagSourceDir, dict)
	if err != nil {
		log.Fatal("Failed to open storage:", err)
	}
	log.Println("Examples loaded:", storage.ExampleCount())

	svc := &sarfyaservice.Service{Dictionary: dict, Storage: storage}

	api, errCh := webapi.Setup(*flagListenAddr)

	webapi.Utils(api.Group("/api/utils"), dict)
	webapi.Examples(api.Group("/api/examples"), svc, storage)
	templfrontend.Endpoints(api.Group(""), svc, storage)

	go func() {
		example, err := storage.FetchExamples(context.Background(), nil, nil)
		if err != nil {
			return
		}

		exists := make(map[string]bool)
		for _, example := range example {
			rt := strings.TrimSpace(example.Text.RawText())
			if exists[rt] {
				log.Println("Duplicate example:", example.ID, example.Text.String())
			}

			exists[rt] = true
		}
	}()

	log.Println("Listening on", *flagListenAddr)

	err = <-errCh
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}
}
