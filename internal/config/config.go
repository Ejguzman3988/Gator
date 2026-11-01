package config

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"html"
	"io"
	"net/http"
	"os"
)

type Config struct {
	DBUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func getConfigFilePath() (string, error) {
	const configFileName = ".gatorconfig.json"
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	filePath := homeDir + "/" + configFileName
	return filePath, nil
}

func Read() (Config, error) {

	filePath, err := getConfigFilePath()

	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(filePath)

	if err != nil {
		return Config{}, err
	}

	config := Config{}

	if err = json.Unmarshal(data, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func (c *Config) SetUser(username string) error {
	c.CurrentUserName = username
	filePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	newData, err := json.Marshal(c)

	if err != nil {
		return err
	}
	os.WriteFile(filePath, newData, 0644)

	return nil
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {

	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)

	req.Header.Set("User-Agent", "gator")

	if err != nil {
		return &RSSFeed{}, err
	}

	res, err := client.Do(req)

	if err != nil {
		return &RSSFeed{}, err
	}

	body, err := io.ReadAll(res.Body)

	defer res.Body.Close()

	data := RSSFeed{}
	err = xml.Unmarshal(body, &data)

	data.Channel.Title = html.UnescapeString(data.Channel.Title)
	data.Channel.Description = html.UnescapeString(data.Channel.Description)

	for idx, item := range data.Channel.Item {
		data.Channel.Item[idx].Title = html.UnescapeString(item.Title)
		data.Channel.Item[idx].Link = html.UnescapeString(item.Link)
		data.Channel.Item[idx].Description = html.UnescapeString(item.Description)
		data.Channel.Item[idx].PubDate = html.UnescapeString(item.PubDate)
	}

	if err != nil {
		return &RSSFeed{}, err
	}

	return &data, nil
}
