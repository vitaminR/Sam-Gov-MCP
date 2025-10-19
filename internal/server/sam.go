package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const samAPIBaseURL = "https://api.sam.gov/prod/opportunities/v2/search"

// Opportunity represents a single SAM.gov opportunity entry returned by the API.
type Opportunity struct {
	Title    string `json:"title"`
	Agency   string `json:"agency"`
	Modified string `json:"modifiedDate"`
	URL      string `json:"url"`
}

// SamSearchResults is the top-level shape from SAM.gov for the opportunities search.
type SamSearchResults struct {
	Results []Opportunity `json:"opportunitiesData"`
}

// SearchOpportunities queries the public SAM.gov opportunities API with the provided params.
func SearchOpportunities(apiKey string, params map[string]interface{}) (*SamSearchResults, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("SAM_API_KEY not set")
	}

	q := url.Values{}
	q.Set("api_key", apiKey)

	if val, ok := params["q"].(string); ok {
		q.Set("q", val)
	}
	if val, ok := params["limit"].(float64); ok {
		q.Set("limit", fmt.Sprintf("%.0f", val))
	}
	// Add other parameters as needed

	req, err := http.NewRequest("GET", samAPIBaseURL+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SAM API request failed with status: %s", resp.Status)
	}

	var results SamSearchResults
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	return &results, nil
}
