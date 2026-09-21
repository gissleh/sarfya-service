package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"sync/atomic"

	"github.com/gissleh/sarfya"
	"github.com/gissleh/sarfya-service/adapters/fwewdictionary"
	"github.com/gissleh/sarfya-service/adapters/sourcestorage"
	"github.com/gissleh/sarfya-service/emphasis"
	"github.com/gissleh/sarfya/adapters/placeholderdictionary"
	"golang.org/x/sync/errgroup"
)

var flagSourceDir = flag.String("source-dir", "./data", "Source directory")
var flagDestFile = flag.String("dest-file", "./stress-data.json", "Destination file")

func main() {
	flag.Parse()

	dict := sarfya.CombinedDictionary{
		sarfya.WithDerivedPoS(fwewdictionary.Global()),
		placeholderdictionary.New(),
	}

	storage, err := sourcestorage.Open(context.Background(), *flagSourceDir, dict)
	if err != nil {
		log.Fatal("Failed to open storage:", err)
	}

	eg := errgroup.Group{}
	eg.SetLimit(8)

	safeCount := uint32(0)
	totalCount := uint32(0)

	examples := storage.AllExamples()
	results := make([]*emphasis.FitResult, len(examples))

	for i, example := range examples {
		if len(example.Text) == 0 {
			continue
		}

		eg.Go(func() error {
			fitRes, err := storage.FindEmphasis(context.Background(), example.ID)
			if err != nil {
				return err
			}

			if fitRes.IsSafe() {
				atomic.AddUint32(&safeCount, 1)
			}
			totalCount := atomic.AddUint32(&totalCount, 1)

			results[i] = fitRes
			if totalCount%100 == 0 {
				log.Println("Saved", totalCount, "example stresses.")
			}

			return nil
		})
	}

	res := make(map[string]emphasis.FitResult, atomic.LoadUint32(&totalCount))
	for i, result := range results {
		if result != nil {
			res[examples[i].ID] = *result
		}
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

	log.Printf("Safe examples: %d/%d (%.2f%%)", safeCount, totalCount, 100*float32(safeCount)/float32(totalCount))
}
