package client

import (
	"encoding/json"
	"fmt"
)

// IrcChannel represents an IRC channel from the daemon.
type IrcChannel struct {
	ChannelID   string `json:"channel_id"`
	Name        string `json:"name"`
	Topic       string `json:"topic"`
	OpenedBy    string `json:"opened_by"`
	Status      int    `json:"status"`
	StatusLabel string `json:"status_label"`
	OpenedAt    int64  `json:"opened_at"`
}

// ListChannels returns available IRC channels from the daemon.
func (c *Client) ListChannels() ([]IrcChannel, error) {
	resp, err := c.get("/api/irc/channels")
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, fmt.Errorf("list channels failed: %s", resp.Error)
	}

	var result struct {
		Channels []IrcChannel `json:"channels"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse channels response: %w", err)
	}

	return result.Channels, nil
}

// OpenChannel creates a new IRC channel.
func (c *Client) OpenChannel(name, topic string) (*IrcChannel, error) {
	body := map[string]interface{}{
		"name": name,
	}
	if topic != "" {
		body["topic"] = topic
	}

	resp, err := c.post("/api/irc/channels/open", body)
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, fmt.Errorf("open channel failed: %s", resp.Error)
	}

	var ch IrcChannel
	if err := json.Unmarshal(resp.Result, &ch); err != nil {
		return nil, fmt.Errorf("failed to parse open channel response: %w", err)
	}

	return &ch, nil
}

// JoinChannel joins an IRC channel's message stream.
func (c *Client) JoinChannel(channelID string) error {
	resp, err := c.post("/api/irc/channels/"+channelID+"/join", nil)
	if err != nil {
		return err
	}

	if !resp.Ok {
		return fmt.Errorf("join channel failed: %s", resp.Error)
	}

	return nil
}

// PartChannel leaves an IRC channel's message stream.
func (c *Client) PartChannel(channelID string) error {
	resp, err := c.post("/api/irc/channels/"+channelID+"/part", nil)
	if err != nil {
		return err
	}

	if !resp.Ok {
		return fmt.Errorf("part channel failed: %s", resp.Error)
	}

	return nil
}

// SendIrcMessage sends a message to an IRC channel.
func (c *Client) SendIrcMessage(channelID, content, nick string) error {
	body := map[string]interface{}{
		"content": content,
		"nick":    nick,
	}

	resp, err := c.post("/api/irc/channels/"+channelID+"/messages", body)
	if err != nil {
		return err
	}

	if !resp.Ok {
		return fmt.Errorf("send message failed: %s", resp.Error)
	}

	return nil
}
