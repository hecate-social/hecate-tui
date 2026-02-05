package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Conversation is a saved chat session.
type Conversation struct {
	ID        string             `json:"id"`
	Title     string             `json:"title"`
	Model     string             `json:"model,omitempty"`
	Messages  []ConversationMsg  `json:"messages"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// ConversationMsg is a single message in a conversation.
type ConversationMsg struct {
	Role    string    `json:"role"`
	Content string    `json:"content"`
	Time    time.Time `json:"time"`
}

// ConversationsDir returns ~/.config/hecate-tui/conversations/.
// Falls back to old path if new dir doesn't exist but old one does.
func ConversationsDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = filepath.Join(os.Getenv("HOME"), ".config")
	}

	newDir := filepath.Join(dir, "hecate-tui", "conversations")
	oldDir := filepath.Join(dir, "hecate", "conversations")

	// If new dir exists (or old doesn't), use new
	if _, err := os.Stat(newDir); err == nil {
		return newDir
	}

	// If old dir exists, migrate it
	if _, err := os.Stat(oldDir); err == nil {
		// Create parent dir and rename
		_ = os.MkdirAll(filepath.Dir(newDir), 0755)
		if os.Rename(oldDir, newDir) == nil {
			return newDir
		}
		// Rename failed — use old dir as fallback
		return oldDir
	}

	// Neither exists — use new path (will be created on first save)
	return newDir
}

// NewConversationID generates a time-based conversation ID.
func NewConversationID() string {
	return time.Now().Format("20060102-150405")
}

// SaveConversation writes a conversation to disk.
func SaveConversation(conv Conversation) error {
	dir := ConversationsDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	conv.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(conv, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(dir, conv.ID+".json")
	return os.WriteFile(path, append(data, '\n'), 0644)
}

// LoadConversation reads a conversation by ID.
func LoadConversation(id string) (Conversation, error) {
	path := filepath.Join(ConversationsDir(), id+".json")

	data, err := os.ReadFile(path)
	if err != nil {
		return Conversation{}, fmt.Errorf("conversation not found: %s", id)
	}

	var conv Conversation
	if err := json.Unmarshal(data, &conv); err != nil {
		return Conversation{}, fmt.Errorf("corrupt conversation file: %w", err)
	}

	return conv, nil
}

// DeleteConversation removes a conversation by ID.
func DeleteConversation(id string) error {
	path := filepath.Join(ConversationsDir(), id+".json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("conversation not found: %s", id)
	}
	return os.Remove(path)
}

// ListConversations returns all saved conversations, newest first.
func ListConversations() []Conversation {
	dir := ConversationsDir()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var convs []Conversation
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		id := strings.TrimSuffix(entry.Name(), ".json")
		conv, err := LoadConversation(id)
		if err != nil {
			continue
		}
		convs = append(convs, conv)
	}

	// Sort newest first
	sort.Slice(convs, func(i, j int) bool {
		return convs[i].UpdatedAt.After(convs[j].UpdatedAt)
	})

	return convs
}

// TitleFromMessages derives a conversation title from the first user message.
func TitleFromMessages(msgs []ConversationMsg) string {
	for _, m := range msgs {
		if m.Role == "user" {
			title := m.Content
			if len(title) > 60 {
				title = title[:57] + "..."
			}
			// Single line
			if idx := strings.IndexByte(title, '\n'); idx >= 0 {
				title = title[:idx]
			}
			return title
		}
	}
	return "Empty conversation"
}
