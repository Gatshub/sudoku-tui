package ui

import "github.com/charmbracelet/lipgloss"

// Palette adaptative (clair / sombre), d'inspiration « Tokyo Night ».
var (
	colBG       = lipgloss.AdaptiveColor{Light: "#e9e9ef", Dark: "#1a1b26"}
	colFG       = lipgloss.AdaptiveColor{Light: "#343b58", Dark: "#c0caf5"}
	colMuted    = lipgloss.AdaptiveColor{Light: "#8a8fa3", Dark: "#565f89"}
	colGiven    = lipgloss.AdaptiveColor{Light: "#1a1b26", Dark: "#c0caf5"}
	colUser     = lipgloss.AdaptiveColor{Light: "#0b5b86", Dark: "#7dcfff"}
	colNote     = lipgloss.AdaptiveColor{Light: "#8a8fa3", Dark: "#565f89"}
	colConflict = lipgloss.AdaptiveColor{Light: "#b00020", Dark: "#f7768e"}
	colAccent   = lipgloss.AdaptiveColor{Light: "#6a4fa3", Dark: "#bb9af7"}
	colBlue     = lipgloss.AdaptiveColor{Light: "#2f5da8", Dark: "#7aa2f7"}
	colGreen    = lipgloss.AdaptiveColor{Light: "#2f6b3f", Dark: "#9ece6a"}
	colYellow   = lipgloss.AdaptiveColor{Light: "#8a5a12", Dark: "#e0af68"}
	colLine     = lipgloss.AdaptiveColor{Light: "#9aa0b0", Dark: "#565f89"}
	colCursorBG = lipgloss.AdaptiveColor{Light: "#b9c6f0", Dark: "#3d59a1"}
	colCursorFG = lipgloss.AdaptiveColor{Light: "#1a1b26", Dark: "#ffffff"}
	colPeerBG   = lipgloss.AdaptiveColor{Light: "#dfe3ef", Dark: "#1f2335"}
	colSameBG   = lipgloss.AdaptiveColor{Light: "#c9d3f4", Dark: "#2b3050"}
)

// Caractères de tracé des blocs 3x3 (lignes épaisses).
const (
	bTopLeft  = "┏"
	bTopRight = "┓"
	bBotLeft  = "┗"
	bBotRight = "┛"
	bTopT     = "┳"
	bBotT     = "┻"
	bCross    = "╋"
	bLeftT    = "┣"
	bRightT   = "┫"
	bHoriz    = "━"
	bVert     = "┃"
)

// Styles de base. Les traits de grille n'ont volontairement pas de fond :
// seules les cases surlignées en portent un, ce qui reste propre même sur un
// terminal à fond clair ou transparent.
var (
	styleLine     = lipgloss.NewStyle().Foreground(colLine)
	styleTitle    = lipgloss.NewStyle().Foreground(colAccent).Bold(true)
	styleSubtitle = lipgloss.NewStyle().Foreground(colMuted)
	styleAccent   = lipgloss.NewStyle().Foreground(colBlue).Bold(true)
	styleGood     = lipgloss.NewStyle().Foreground(colGreen).Bold(true)
	styleWarn     = lipgloss.NewStyle().Foreground(colYellow)
	styleErr      = lipgloss.NewStyle().Foreground(colConflict).Bold(true)
	styleMuted    = lipgloss.NewStyle().Foreground(colMuted)
	styleKey      = lipgloss.NewStyle().Foreground(colAccent)
	styleHelpKey  = lipgloss.NewStyle().Foreground(colBlue).Bold(true)
	styleSelItem  = lipgloss.NewStyle().Foreground(colBG).Background(colBlue).Bold(true).Padding(0, 2)
	styleMenuItem = lipgloss.NewStyle().Foreground(colFG).Padding(0, 2)
	styleBox      = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colAccent).
			Padding(1, 3)
)

// styleCell renvoie le style d'une case selon son fond et sa couleur de texte.
func styleCell(w int, bg lipgloss.TerminalColor, fg lipgloss.TerminalColor) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(w).
		Align(lipgloss.Center).
		Background(bg).
		Foreground(fg)
}
