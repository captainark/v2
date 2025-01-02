// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Discord Webhooks documentation: https://discord.com/developers/docs/resources/webhook

package discord // import "miniflux.app/v2/internal/integration/discord"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second

type Client struct {
	webhookURL string
}

func NewClient(webhookURL string) *Client {
	return &Client{webhookURL: webhookURL}
}

func (c *Client) SendDiscordMsg(feed *model.Feed, entries model.Entries) error {

	for _, entry := range entries {

		footerText := entry.Author + " " + "•" + " " + "Miniflux/" +version.Version

		requestBody, err := json.Marshal(&discordMessage{
			Embeds: []discordEmbed{
				{
					Title: entry.Title,
					Url: entry.URL,
					Description: feed.Title,
					Color: 5793266,
					Footer: &discordFooter{
						Text: footerText,
						IconUrl: feed.IconURL,
					},
				},
			},
		})
		if err != nil {
			return fmt.Errorf("discord: unable to encode request body: %v", err)
		}

		request, err := http.NewRequest(http.MethodPost, c.webhookURL, bytes.NewReader(requestBody))
		if err != nil {
			return fmt.Errorf("discord: unable to create request: %v", err)
		}

		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("User-Agent", "Miniflux/"+version.Version)

		slog.Debug("Sending Discord notification",
			slog.String("webhookURL", c.webhookURL),
			slog.String("title", feed.Title),
			slog.String("entry_url", entry.URL),
		)

		httpClient := &http.Client{Timeout: defaultClientTimeout}
		response, err := httpClient.Do(request)
		if err != nil {
			return fmt.Errorf("discord: unable to send request: %v", err)
		}
		defer response.Body.Close()

		if response.StatusCode >= 400 {
			return fmt.Errorf("discord: unable to send a notification: url=%s status=%d", c.webhookURL, response.StatusCode)
		}
	}

	return nil
}

type discordFooter struct {
	Text    string `json:"text,omitempty"`
	IconUrl string `json:"icon_url,omitempty"`
}

type discordEmbed struct {
	Title       string           `json:"title,omitempty"`
	Url         string           `json:"url,omitempty"`
	Description string           `json:"description,omitempty"`
	Color       int              `json:"color,omitempty"`
	Footer      *discordFooter  `json:"footer,omitempty"`
}

type discordMessage struct {
	Username        string          `json:"username,omitempty"`
	AvatarUrl       string          `json:"avatar_url,omitempty"`
	Content         string          `json:"content,omitempty"`
	Embeds          []discordEmbed  `json:"embeds,omitempty"`
}
