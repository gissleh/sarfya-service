package templfrontend

import (
	"context"
	"math/rand/v2"
	"strings"
	"sync"
	"time"

	"github.com/gissleh/sarfya"
	"github.com/gissleh/sarfya-service/emphasis"
	"github.com/gissleh/sarfya/sarfyaservice"
)

var exampleCount = 0

var demoCtxKey = &struct{ key string }{key: "demoCtxKey"}
var demoEmphasisCtxKey = &struct{ key string }{key: "demoEmphasisCtxKey"}
var langCtxKey = &struct{ key string }{key: "langCtxKey"}

func findLanguage(match sarfya.FilterMatch, langs string) string {
	for l := range strings.SplitSeq(langs, ",") {
		l := strings.TrimSpace(l)

		if match.Translations[l] != nil {
			return l
		}
	}

	return ""
}

var eotdMutex sync.Mutex
var eotd *sarfya.Example
var eotdEmphasis *emphasis.FitResult
var eotdAt time.Time

func getEotd(storage sarfyaservice.ExampleStorage, emphasisStorage emphasis.Storage) (*sarfya.Example, *emphasis.FitResult) {
	eotdMutex.Lock()
	defer eotdMutex.Unlock()

	if eotd != nil && time.Now().Truncate(24*time.Hour).Equal(eotdAt) {
		return eotd, eotdEmphasis
	}

	examples, _ := storage.FetchExamples(context.Background(), nil, nil)
	rng := rand.NewPCG(uint64(time.Now().Truncate(time.Hour*24).Unix()), uint64(time.Now().Truncate(time.Hour*24).Unix()))

	index := int(rng.Uint64() % uint64(len(examples)))
	for i := range len(examples) {
		i := (index + i) % len(examples)
		if len(examples[i].Translations["en"]) == 0 {
			continue
		}
		if strings.HasPrefix(examples[i].Translations["en"][0].RawText(), "(") {
			continue
		}

		exampleEmphasis, err := emphasisStorage.FindEmphasis(context.Background(), examples[i].ID)
		if err != nil || !exampleEmphasis.IsSafe() {
			continue
		}

		eotd = new(examples[i])
		eotdEmphasis = exampleEmphasis
		eotdAt = time.Now().Truncate(24 * time.Hour)

		return eotd, eotdEmphasis
	}

	return new(examples[index]), nil
}
