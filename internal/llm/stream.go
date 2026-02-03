package llm

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

// StreamParser parses SSE/NDJSON streams from the LLM API
type StreamParser struct {
	reader *bufio.Reader
}

// NewStreamParser creates a new stream parser
func NewStreamParser(r io.Reader) *StreamParser {
	return &StreamParser{
		reader: bufio.NewReader(r),
	}
}

// Next reads the next response chunk from the stream
// Returns nil, io.EOF when the stream is exhausted
func (p *StreamParser) Next() (*ChatResponse, error) {
	for {
		line, err := p.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Handle SSE format: "data: {...}"
		if strings.HasPrefix(line, "data:") {
			line = strings.TrimPrefix(line, "data:")
			line = strings.TrimSpace(line)
		}

		// Skip SSE comments
		if strings.HasPrefix(line, ":") {
			continue
		}

		// Parse JSON
		var resp ChatResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			// Skip malformed lines
			continue
		}

		return &resp, nil
	}
}

// ParseAll reads all chunks from the stream
func (p *StreamParser) ParseAll() ([]ChatResponse, error) {
	var responses []ChatResponse
	for {
		resp, err := p.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return responses, err
		}
		responses = append(responses, *resp)
	}
	return responses, nil
}
