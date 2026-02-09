package chat

import (
	"regexp"
	"strings"
)

var thinkTagRe = regexp.MustCompile(`(?s)<think>(.*?)</think>`)

// StripThinkTags removes <think>...</think> blocks from text and returns
// the visible content and the extracted thinking content separately.
func StripThinkTags(text string) (visible string, thinking string) {
	matches := thinkTagRe.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return text, ""
	}

	var thinkParts []string
	for _, m := range matches {
		thinkParts = append(thinkParts, strings.TrimSpace(m[1]))
	}

	visible = thinkTagRe.ReplaceAllString(text, "")
	visible = strings.TrimSpace(visible)
	thinking = strings.Join(thinkParts, "\n\n")
	return visible, thinking
}

// HasThinkTags returns true if the text contains any <think> tags.
func HasThinkTags(text string) bool {
	return strings.Contains(text, "<think>")
}

// HasOpenThinkTag returns true if the text has an unclosed <think> tag.
// Used during streaming to detect partial think blocks.
func HasOpenThinkTag(text string) bool {
	opens := strings.Count(text, "<think>")
	closes := strings.Count(text, "</think>")
	return opens > closes
}

// SplitAtOpenThink splits text at the last unclosed <think> tag,
// returning the content before it and the partial think content after it.
// If no open tag exists, returns the full text and empty string.
func SplitAtOpenThink(text string) (before string, partial string) {
	// Find the last <think> that has no matching </think> after it
	lastOpen := strings.LastIndex(text, "<think>")
	if lastOpen == -1 {
		return text, ""
	}

	// Check if there's a </think> after this last <think>
	after := text[lastOpen:]
	if strings.Contains(after, "</think>") {
		// Fully closed, no partial
		return text, ""
	}

	before = text[:lastOpen]
	partial = text[lastOpen+len("<think>"):]
	return before, partial
}
