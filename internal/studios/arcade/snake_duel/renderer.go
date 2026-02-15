package snake_duel

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Cell types for the grid.
type cellType int

const (
	cellEmpty cellType = iota
	cellSnake1Body
	cellSnake1Head
	cellSnake2Body
	cellSnake2Head
	cellFood
	cellPoison
)

// Color palette — true color hex values.
var (
	colorBg         = lipgloss.Color("#0f0f1a")
	colorGridDot    = lipgloss.Color("#2a2a50")
	colorSnake1Body = lipgloss.Color("#3b82f6")
	colorSnake1Head = lipgloss.Color("#60a5fa")
	colorSnake2Body = lipgloss.Color("#ef4444")
	colorSnake2Head = lipgloss.Color("#f87171")
	colorFood       = lipgloss.Color("#fbbf24")
	colorPoison     = lipgloss.Color("#a855f7")
)

// cellColor returns the foreground color for a cell type.
func cellColor(c cellType) lipgloss.Color {
	switch c {
	case cellSnake1Body:
		return colorSnake1Body
	case cellSnake1Head:
		return colorSnake1Head
	case cellSnake2Body:
		return colorSnake2Body
	case cellSnake2Head:
		return colorSnake2Head
	case cellFood:
		return colorFood
	case cellPoison:
		return colorPoison
	default:
		return colorGridDot
	}
}

// RenderGrid builds the half-block grid from a GameState.
// 30 columns x 12 terminal rows (each row = 2 vertical game cells).
func RenderGrid(gs GameState) string {
	// Build 30x24 cell grid
	var grid [GridWidth][GridHeight]cellType

	// Place snake1 body + head
	for i, pt := range gs.Snake1.Body {
		x, y := pt[0], pt[1]
		if inBounds(x, y) {
			if i == 0 {
				grid[x][y] = cellSnake1Head
			} else {
				grid[x][y] = cellSnake1Body
			}
		}
	}

	// Place snake2 body + head
	for i, pt := range gs.Snake2.Body {
		x, y := pt[0], pt[1]
		if inBounds(x, y) {
			if i == 0 {
				grid[x][y] = cellSnake2Head
			} else {
				grid[x][y] = cellSnake2Body
			}
		}
	}

	// Place food
	fx, fy := gs.Food[0], gs.Food[1]
	if inBounds(fx, fy) {
		grid[fx][fy] = cellFood
	}

	// Place poison apples
	for _, p := range gs.PoisonApples {
		px, py := p.Pos[0], p.Pos[1]
		if inBounds(px, py) {
			grid[px][py] = cellPoison
		}
	}

	// Render pairs of rows using half-block characters
	// Each terminal row represents 2 game rows (top + bottom)
	var rows []string
	termRows := GridHeight / 2
	for tr := 0; tr < termRows; tr++ {
		topY := tr * 2
		botY := topY + 1
		var row strings.Builder
		for x := 0; x < GridWidth; x++ {
			top := grid[x][topY]
			bot := grid[x][botY]
			row.WriteString(renderHalfBlock(top, bot))
		}
		rows = append(rows, row.String())
	}

	// Wrap in a thin border
	content := strings.Join(rows, "\n")
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3a3a5c")).
		Background(colorBg)

	return borderStyle.Render(content)
}

// renderHalfBlock returns a styled character for a top/bottom cell pair.
//
// Unicode half-block rendering:
//   - ▀ (U+2580): top half — fg = top color, bg = bottom color
//   - ▄ (U+2584): bottom half — fg = bottom color, bg = top color
//   - █ (U+2588): full block — both same color
//   - · (dot): empty grid position
//   - (space): both empty
func renderHalfBlock(top, bot cellType) string {
	topEmpty := top == cellEmpty
	botEmpty := bot == cellEmpty

	switch {
	case topEmpty && botEmpty:
		// Both empty — show dim grid dot
		return lipgloss.NewStyle().
			Foreground(colorGridDot).
			Background(colorBg).
			Render("·")

	case !topEmpty && botEmpty:
		// Top occupied, bottom empty
		return lipgloss.NewStyle().
			Foreground(cellColor(top)).
			Background(colorBg).
			Render("▀")

	case topEmpty && !botEmpty:
		// Bottom occupied, top empty
		return lipgloss.NewStyle().
			Foreground(cellColor(bot)).
			Background(colorBg).
			Render("▄")

	default:
		// Both occupied
		if top == bot {
			// Same type — full block
			return lipgloss.NewStyle().
				Foreground(cellColor(top)).
				Background(cellColor(top)).
				Render("█")
		}
		// Different types — upper half block with fg=top, bg=bottom
		return lipgloss.NewStyle().
			Foreground(cellColor(top)).
			Background(cellColor(bot)).
			Render("▀")
	}
}

func inBounds(x, y int) bool {
	return x >= 0 && x < GridWidth && y >= 0 && y < GridHeight
}
