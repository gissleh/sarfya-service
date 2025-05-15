package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/gissleh/sarfya"
	"github.com/gissleh/sarfya-service/adapters/fwewdictionary"
	"github.com/gissleh/sarfya-service/adapters/sourcestorage"
	"github.com/gissleh/sarfya-service/emphasis"
	"github.com/gissleh/sarfya/adapters/placeholderdictionary"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"sync"
)

var flagSourceDir = flag.String("source-dir", "./data", "Source directory")
var flagDestFile = flag.String("dest-file", "./stress-data.json", "Destination file")
var flagLitxapApi = flag.String("litxap-api", "http://localhost:8081", "Litxap root address")

func main() {
	flag.Parse()

	dict := sarfya.CombinedDictionary{
		sarfya.WithDerivedPoS(fwewdictionary.Global()),
		placeholderdictionary.New(),
	}

	storage, err := sourcestorage.Open(context.Background(), *flagLitxapApi, *flagSourceDir, dict)
	if err != nil {
		log.Fatal("Failed to open storage:", err)
	}

	resMu := sync.Mutex{}
	res := make(map[string]emphasis.FitResult, 2048)

	eg := errgroup.Group{}
	eg.SetLimit(4)

	for _, example := range storage.AllExamples() {
		if len(example.Text) == 0 {
			continue
		}

		eg.Go(func() error {
			fitRes, err := storage.FindEmphasis(context.Background(), example.ID)
			if err != nil {
				return err
			}

			resMu.Lock()
			res[example.ID] = *fitRes
			if len(res)%100 == 0 {
				log.Println("Saved", len(res), "example stresses.")
			}
			resMu.Unlock()

			return nil
		})
	}

	err = eg.Wait()
	if err != nil {
		log.Fatal("Litxap failed:", err)
	}

	f, err := os.OpenFile(*flagDestFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		log.Fatal("Failed to open destination file:", err)
	}

	err = json.NewEncoder(f).Encode(res)
	if err != nil {
		log.Fatal("Failed to encode results:", err)
	}

	err = f.Close()
	if err != nil {
		log.Fatal("Failed to close destination file:", err)
	}
}
