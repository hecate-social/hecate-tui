package app

import (
	"github.com/atotto/clipboard"
	"github.com/hecate-social/hecate-tui/internal/chat"
	"github.com/hecate-social/hecate-tui/internal/config"

	tea "github.com/charmbracelet/bubbletea"
)

func (a *App) saveConversation() {
	msgs := a.chat.Messages()
	if len(msgs) == 0 {
		return
	}

	var convMsgs []config.ConversationMsg
	for _, m := range msgs {
		if m.Role == "system" {
			continue // Don't persist command output
		}
		convMsgs = append(convMsgs, config.ConversationMsg{
			Role:    m.Role,
			Content: m.Content,
			Time:    m.Time,
		})
	}

	if len(convMsgs) == 0 {
		return
	}

	title := config.TitleFromMessages(convMsgs)
	a.conversationTitle = title

	conv := config.Conversation{
		ID:        a.conversationID,
		Title:     title,
		Model:     a.chat.ActiveModelName(),
		Messages:  convMsgs,
		CreatedAt: convMsgs[0].Time,
	}

	if err := config.SaveConversation(conv); err != nil {
		a.chat.InjectSystemMessage("Warning: failed to save conversation: " + err.Error())
	}
}

func (a *App) startNewConversation() {
	a.saveConversation()
	a.chat.ClearMessages()
	a.conversationID = config.NewConversationID()
	a.conversationTitle = ""
}

func (a *App) loadConversation(id string) error {
	conv, err := config.LoadConversation(id)
	if err != nil {
		return err
	}

	a.saveConversation() // save current first

	var msgs []chat.Message
	for _, m := range conv.Messages {
		msgs = append(msgs, chat.Message{
			Role:    m.Role,
			Content: m.Content,
			Time:    m.Time,
		})
	}

	a.chat.ClearMessages()
	a.chat.LoadMessages(msgs)
	a.conversationID = conv.ID
	a.conversationTitle = conv.Title
	return nil
}

func (a *App) yankLastResponse() tea.Cmd {
	content := a.chat.LastAssistantMessage()
	if content == "" {
		a.chat.InjectSystemMessage("No response to copy.")
		return nil
	}

	if err := clipboard.WriteAll(content); err != nil {
		a.chat.InjectSystemMessage("Clipboard unavailable: " + err.Error())
		return nil
	}

	// Truncate preview
	preview := content
	if len(preview) > 60 {
		preview = preview[:57] + "..."
	}
	a.chat.InjectSystemMessage("Copied to clipboard: " + preview)
	return nil
}
