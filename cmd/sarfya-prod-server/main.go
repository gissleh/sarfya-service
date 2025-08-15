package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gissleh/sarfya"
	"github.com/gissleh/sarfya-service/adapters/fwewdictionary"
	"github.com/gissleh/sarfya-service/adapters/jsonemphasisdata"
	"github.com/gissleh/sarfya-service/adapters/templfrontend"
	"github.com/gissleh/sarfya-service/adapters/webapi"
	"github.com/gissleh/sarfya/adapters/jsonstorage"
	"github.com/gissleh/sarfya/adapters/placeholderdictionary"
	"github.com/gissleh/sarfya/sarfyaservice"
)

var flagSourceFile = flag.String("source-file", "./data-compiled.json", "File containing data.")
var flagListenAddr = flag.String("listen", ":$PORT", "Listen address")
var flagEmphasisFile = flag.String("emphasis-file", "./stress-data.json", "File containing stress data.")

func main() {
	dict := sarfya.CombinedDictionary{
		sarfya.WithDerivedPoS(fwewdictionary.Global()),
		placeholderdictionary.New(),
	}

	if port := os.Getenv("PORT"); port != "" {
		*flagListenAddr = strings.Replace(*flagListenAddr, "$PORT", port, 1)
	} else {
		*flagListenAddr = strings.Replace(*flagListenAddr, "$PORT", "8080", 1)
	}

	storage, err := jsonstorage.Open(*flagSourceFile, true)
	if err != nil {
		log.Fatalln("Failed to open json storage:", err)
		return
	}

	svc := &sarfyaservice.Service{Dictionary: dict, Storage: storage, ReadOnly: true}
	api, errCh := webapi.Setup(*flagListenAddr)

	api.File("/data.json", *flagSourceFile)

	emphasisStorage, err := jsonemphasisdata.Load(*flagEmphasisFile)
	if err != nil {
		log.Fatalln("Failed to load emphasis:", err)
		return
	}

	webapi.Utils(api.Group("/api/utils"), dict)
	webapi.Examples(api.Group("/api/examples"), svc, emphasisStorage)
	templfrontend.Endpoints(api.Group(""), svc, emphasisStorage)

	log.Println("Listening on", *flagListenAddr)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-signalCh:
		log.Println("Shutting down due to signal:", sig)
	case err := <-errCh:
		log.Fatal("Failed to listen:", err)
	}
}
