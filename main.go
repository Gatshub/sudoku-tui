// Commande sudoku-tui : un sudoku pour le terminal, bâti avec Bubble Tea
// et Lip Gloss (projet Charm).
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"sudoku-tui/ui"
)

func main() {
	p := tea.NewProgram(
		ui.New(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "sudoku-tui :", err)
		os.Exit(1)
	}
}
