package reports

import "net/http"

const baseUrl = "https://botw-compendium.herokuapp.com/api/v3/compendium"

type HttpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	httpClient HttpClient
	baseURL    string
}

func NewClient(baseUrl string, httpClient HttpClient) *Client {
	if baseUrl == "" {
		baseUrl = "https://botw-compendium.herokuapp.com/api/v3/compendium"
	}
	return &Client{
		httpClient: httpClient,
		baseURL:    baseUrl,
	}
}

type Monster struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Id 		int   `json:"id"`
	Category    string `json:"category"`
	CommonLocations []string `json:"common_locations"`
	Drops	   []string `json:"drops"`
	Dlc bool `json:"dlc"`
}