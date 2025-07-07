package notify

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var slackAPIURL = "https://slack.com/api/chat.postMessage"

// SlackNotifier sends messages to Slack via webhook or OAuth token.
type SlackNotifier struct {
	webhookURL string
	token      string
	channel    string
	client     *http.Client
}

// NewSlackNotifier creates a notifier. At least one of webhookURL or token must be provided.
func NewSlackNotifier(webhookURL, token, channel string) *SlackNotifier {
	return &SlackNotifier{
		webhookURL: webhookURL,
		token:      token,
		channel:    channel,
		client:     &http.Client{Timeout: 5 * time.Second},
	}
}

// SendMessage posts a text message to Slack.
func (s *SlackNotifier) SendMessage(text string) error {
	if s.webhookURL != "" {
		payload := map[string]string{"text": text}
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(b))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= http.StatusMultipleChoices {
			return fmt.Errorf("slack webhook failed: %s", resp.Status)
		}
		return nil
	}

	if s.token == "" {
		return errors.New("no slack configuration provided")
	}

	payload := map[string]string{"channel": s.channel, "text": text}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", slackAPIURL, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("slack api error: %s", resp.Status)
	}
	return nil
}
