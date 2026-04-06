package genius

import (
	"context"
	"drpp/server/logger"
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const searchEndpoint = "https://api.genius.com/search"

type searchResponse struct {
	Response struct {
		Hits []struct {
			Result struct {
				Url string `json:"url"`
			} `json:"result"`
		} `json:"hits"`
	} `json:"response"`
}

// Search queries the Genius API for the given artist and title and returns the
// URL of the first matching result. If the API key is empty, a browser search
// URL is returned instead.
func Search(ctx context.Context, apiKey string, artist string, title string) string {
	query := strings.TrimSpace(artist + " " + title)
	if query == "" {
		return ""
	}

	if apiKey == "" {
		return "https://genius.com/search?q=" + url.QueryEscape(query)
	}

	reqUrl := searchEndpoint + "?q=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl, nil)
	if err != nil {
		logger.Error(err, "Failed to create Genius search request")
		return fallbackSearchUrl(query)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error(err, "Failed to execute Genius search request")
		return fallbackSearchUrl(query)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error(nil, "Genius API returned status %d", resp.StatusCode)
		return fallbackSearchUrl(query)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err, "Failed to read Genius search response")
		return fallbackSearchUrl(query)
	}

	var searchResp searchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		logger.Error(err, "Failed to unmarshal Genius search response")
		return fallbackSearchUrl(query)
	}

	if len(searchResp.Response.Hits) == 0 {
		logger.Debug("No Genius results found for %q", query)
		return fallbackSearchUrl(query)
	}

	resultUrl := searchResp.Response.Hits[0].Result.Url
	if resultUrl == "" {
		return fallbackSearchUrl(query)
	}
	return resultUrl
}

func fallbackSearchUrl(query string) string {
	return fmt.Sprintf("https://genius.com/search?q=%s", url.QueryEscape(query))
}
