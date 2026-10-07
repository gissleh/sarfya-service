package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	fwew "github.com/fwew/fwew-lib/v5"
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
var flagListenAddr = flag.String("listen", "127.0.0.1:45678", "Listen address")
var flagEmphasisFile = flag.String("emphasis-file", "./stress-data.json", "File containing stress data.")
var flagOutputDirectory = flag.String("output-directory", "./prerendered-pages", "File containing stress data.")

func main() {
	dict := sarfya.CombinedDictionary{
		sarfya.WithDerivedPoS(fwewdictionary.Global()),
		placeholderdictionary.New(),
	}

	err := os.MkdirAll(*flagOutputDirectory, 0755)
	if err != nil {
		log.Fatal("Failed to create directory:", err)
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

	templfrontend.Endpoints(api.Group(""), svc, emphasisStorage, "")

	log.Println("Listening on", *flagListenAddr)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	doneCh := make(chan error)

	examples, err := svc.Storage.FetchExamples(context.Background(), nil, nil)
	if err != nil {
		return
	}

	go func() {
		entries, err := fwew.List([]string{}, 1)
		if err != nil {
			doneCh <- err
			return
		}

		extras := make([]string, 0, 64)
		for i, entry := range entries {
			navi := entry.Navi
			lookup := navi + ":" + entry.ID
			size, err := runJob(lookup)
			if err != nil {
				doneCh <- err
				return
			}
			if size > 3072 || strings.Contains(entry.PartOfSpeech, "adp.") {
				extras = append(extras, navi)
			}
			if withoutPlus, ok := strings.CutSuffix(navi, "+"); ok {
				extras = append(extras, withoutPlus+":"+entry.ID, withoutPlus)
			}
			if strings.ToLower(navi) != navi {
				extras = append(extras, strings.ToLower(navi)+":"+entry.ID)
				if size > 3072 {
					extras = append(extras, strings.ToLower(navi))
				}
			}

			if (i+1)%100 == 0 {
				log.Printf("Pages pre-rendered: %d/%d", i+1, len(entries))
			}
		}
		log.Printf("Pages pre-rendered: %d/%d", len(entries), len(entries))

		for _, l := range "ABCDEFGHIJKLMNOPQRSTUWXYZ" {
			lookup1 := fmt.Sprintf("%c:P%c", l, l)
			lookup2 := fmt.Sprintf("%c", l)
			extras = append(extras, lookup1, lookup2)
		}

		sourceSeen := make(map[string]bool)
		flagSeen := make(map[sarfya.ExampleFlag]bool)
		for _, example := range examples {
			for _, exampleFlag := range example.Flags {
				if !flagSeen[exampleFlag] {
					flagSeen[exampleFlag] = true
					extras = append(extras, "flag:"+string(exampleFlag))
				}
			}

			if !sourceSeen[example.Source.ID] {
				sourceSeen[example.Source.ID] = true
				extras = append(extras, "src:"+example.Source.ID)
			}
		}

		for i, extra := range extras {
			_, err := runJob(extra)
			if err != nil {
				doneCh <- err
				return
			}

			if (i+1)%100 == 0 {
				log.Printf("Extra pages pre-rendered: %d/%d", i+1, len(extras))
			}
		}
		log.Printf("Extra pages pre-rendered: %d/%d", len(extras), len(extras))
		log.Println("Extras:", strings.Join(extras, ", "))

		close(doneCh)
	}()

	select {
	case sig := <-signalCh:
		log.Println("Shutting down due to signal:", sig)
	case err := <-errCh:
		log.Fatal("Failed to listen:", err)
	case err := <-doneCh:
		if err != nil {
			log.Fatal("Failed to generate pages:", err)
		}
		log.Println("Done!")
	}
}

func runJob(lookup string) (int, error) {
	res, err := http.Get(fmt.Sprintf("http://%s/search/%s", *flagListenAddr, url.PathEscape(lookup)))
	if err != nil {
		return 0, err
	}

	if res.StatusCode != 200 {
		log.Printf("%s - %s", lookup, res.Status)
		_ = res.Body.Close()
		return 0, nil
	}

	data, err := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if err != nil {
		log.Printf("%s - %err: s", lookup, err)
		return 0, nil
	}

	filename := base64.RawURLEncoding.EncodeToString([]byte(lookup)) + ".html.gz"
	filepath := path.Join(*flagOutputDirectory, filename)

	buf := bytes.NewBuffer(make([]byte, 0, 64*1024))
	gzipWriter, err := gzip.NewWriterLevel(buf, gzip.BestCompression)
	if err != nil {
		return 0, err
	}
	gzipWriter.Name = filename[:len(filename)-len(".gz")]
	n, err := gzipWriter.Write(data)
	if err != nil {
		return 0, err
	}
	if n != len(data) {
		return 0, fmt.Errorf("%d/%d bytes written", n, len(data))
	}
	err = gzipWriter.Close()
	if err != nil {
		return 0, err
	}

	err = os.WriteFile(filepath, buf.Bytes(), 0600)
	if err != nil {
		return 0, err
	}

	return buf.Len(), nil
}
