package emphasis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gissleh/litxap"
	"net/http"
	"net/url"
	"strings"
)

type LitxapClient struct {
	ApiURL string
}

func (client *LitxapClient) Call(ctx context.Context, raw string, customWords []string) ([]litxap.Line, error) {
	var resLines []litxap.Line
	for _, line := range strings.Split(strings.Trim(raw, "\n"), "\n") {
		reqUrl := fmt.Sprint(client.ApiURL, "/api/run?line=", url.QueryEscape(line))
		if customWords != nil {
			reqUrl += "&names=" + url.QueryEscape(strings.Join(customWords, ","))
		}

		req, err := http.NewRequestWithContext(ctx, "GET", reqUrl, nil)
		if err != nil {
			return nil, err
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		var response struct {
			Line litxap.Line `json:"line"`
		}
		err = json.NewDecoder(res.Body).Decode(&response)
		_ = res.Body.Close()
		if err != nil {
			return nil, err
		}

		resLines = append(resLines, response.Line)
	}

	return resLines, nil
}
