package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"sudoku-tui/sudoku"
)

func TestCursorMovement(t *testing.T) {
	m := testModel(120, 50)
	m = m.start(sudoku.Difficulties[0])
	m.cursorR, m.cursorC = 4, 4

	step := func(msg tea.Msg) {
		nm, _ := m.Update(msg)
		m = nm.(Model)
	}

	step(runeKey('l'))
	if m.cursorC != 5 {
		t.Fatalf("'l' : colonne %d, attendu 5", m.cursorC)
	}
	step(runeKey('h'))
	if m.cursorC != 4 {
		t.Fatalf("'h' : colonne %d, attendu 4", m.cursorC)
	}
	step(runeKey('j'))
	if m.cursorR != 5 {
		t.Fatalf("'j' : ligne %d, attendu 5", m.cursorR)
	}
	step(runeKey('k'))
	if m.cursorR != 4 {
		t.Fatalf("'k' : ligne %d, attendu 4", m.cursorR)
	}
	step(tea.KeyMsg{Type: tea.KeyRight})
	if m.cursorC != 5 {
		t.Fatalf("→ : colonne %d, attendu 5", m.cursorC)
	}
	step(tea.KeyMsg{Type: tea.KeyLeft})
	if m.cursorC != 4 {
		t.Fatalf("← : colonne %d, attendu 4", m.cursorC)
	}

	// Les bords ne doivent pas déborder.
	m.cursorR, m.cursorC = 0, 0
	step(runeKey('h'))
	step(runeKey('k'))
	if m.cursorR != 0 || m.cursorC != 0 {
		t.Fatalf("débordement en haut à gauche : (%d,%d)", m.cursorR, m.cursorC)
	}
	m.cursorR, m.cursorC = 8, 8
	step(runeKey('l'))
	step(runeKey('j'))
	if m.cursorR != 8 || m.cursorC != 8 {
		t.Fatalf("débordement en bas à droite : (%d,%d)", m.cursorR, m.cursorC)
	}
}

// Une case remplie ne peut pas recevoir de note, et les notes ne remplissent
// pas la case.
func TestNotesBlockedOnFilledCell(t *testing.T) {
	m := testModel(120, 50)
	m = m.start(sudoku.Difficulties[0])

	// Trouver une case vide bien au centre.
	var r, c int
	found := false
	for rr := 3; rr < 6 && !found; rr++ {
		for cc := 3; cc < 6; cc++ {
			if m.board[rr][cc] == 0 {
				r, c, found = rr, cc, true
				break
			}
		}
	}
	if !found {
		t.Skip("pas de case vide au centre")
	}
	m.cursorR, m.cursorC = r, c
	want := m.solution[r][c]

	nm, _ := m.Update(runeKey(rune('0' + want)))
	m = nm.(Model)
	if m.board[r][c] != want {
		t.Fatalf("valeur %d non placée en (%d,%d)", want, r, c)
	}

	nm, _ = m.Update(runeKey('n'))
	m = nm.(Model)
	nm, _ = m.Update(runeKey('1'))
	m = nm.(Model)
	if m.notes[r][c] != 0 {
		t.Fatal("des notes ont été posées sur une case déjà remplie")
	}
	if m.board[r][c] != want {
		t.Fatal("la valeur a été altérée par le mode notes")
	}
}
