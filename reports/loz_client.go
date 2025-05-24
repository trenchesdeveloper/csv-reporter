package reports

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HttpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type LozClient struct {
	httpClient HttpClient
	baseURL    string
}

func NewClient(httpClient HttpClient) *LozClient {

	baseUrl := "https://botw-compendium.herokuapp.com/api/v3/compendium"

	return &LozClient{
		httpClient: httpClient,
		baseURL:    baseUrl,
	}
}

type Monster struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Image           string   `json:"image"`
	Id              int      `json:"id"`
	Category        string   `json:"category"`
	CommonLocations []string `json:"common_locations"`
	Drops           []string `json:"drops"`
	Dlc             bool     `json:"dlc"`
}

type MonstersResponse struct {
	Data []Monster `json:"data"`
}

func (c *LozClient) GetMonsters() (*MonstersResponse, error) {
	url := c.baseURL + "/category/monsters"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	reqUrl := req.URL
	queryParams := req.URL.Query()
	queryParams.Set("game", "totk")
	reqUrl.RawQuery = queryParams.Encode()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var monstersResponse MonstersResponse
	if err := json.NewDecoder(resp.Body).Decode(&monstersResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &monstersResponse, nil
}
