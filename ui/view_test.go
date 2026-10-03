package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"sudoku-tui/sudoku"
)

func testModel(w, h int) Model {
	m := New()
	m.width, m.height = w, h
	m.ready = true
	m.aspect = 2.0 // valeur fixe : les tests doivent être déterministes
	return m
}

func runeKey(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func TestMenuRenders(t *testing.T) {
	out := testModel(100, 40).View()
	if strings.TrimSpace(out) == "" {
		t.Fatal("l'écran menu est vide")
	}
	if !strings.Contains(out, "S U D O K U") {
		t.Fatal("le titre est absent du menu")
	}
}

// La grille doit être parfaitement alignée : toutes les lignes de largeur égale.
func TestGridAligned(t *testing.T) {
	for _, ch := range []int{1, 3} {
		m := testModel(120, 50)
		m = m.start(sudoku.Difficulties[0])
		m.ch = ch
		lines := m.renderGrid()
		if len(lines) != m.gridH() {
			t.Fatalf("ch=%d : %d lignes, attendu %d", ch, len(lines), m.gridH())
		}
		for i, l := range lines {
			if got := lipgloss.Width(l); got != m.gridW() {
				t.Fatalf("ch=%d ligne %d : largeur %d, attendu %d", ch, i, got, m.gridW())
			}
		}
	}
}

// Un clic au centre d'une case doit renvoyer cette case.
func TestCellAtRoundTrip(t *testing.T) {
	m := testModel(100, 40)
	m = m.start(sudoku.Difficulties[0])
	lines, cw, ch := m.fitPlay()
	mm := m
	mm.cw, mm.ch = cw, ch
	top := (m.height - len(lines)) / 2
	gridY := top + 2
	gridX := (m.width - mm.gridW()) / 2

	// Coordonnées locales de chaque case (0 = bord gauche/haut).
	localX := func(c int) int {
		x := 1 + c*cw
		if c >= 3 {
			x++
		}
		if c >= 6 {
			x++
		}
		return x
	}
	localY := func(r int) int {
		y := 1 + r*ch
		if r >= 3 {
			y++
		}
		if r >= 6 {
			y++
		}
		return y
	}

	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			x := gridX + localX(c) + cw/2
			y := gridY + localY(r) + ch/2
			gr, gc, ok := m.cellAt(x, y)
			if !ok {
				t.Fatalf("clic sur la case (%d,%d) refusé", r, c)
			}
			if gr != r || gc != c {
				t.Fatalf("clic (%d,%d) → (%d,%d), attendu (%d,%d)", x, y, gr, gc, r, c)
			}
		}
	}
	// Hors grille : refusé.
	if _, _, ok := m.cellAt(0, 0); ok {
		t.Fatal("un clic hors grille devrait être refusé")
	}
}

func TestPlaceUndoAndWin(t *testing.T) {
	m := testModel(120, 50)
	m = m.start(sudoku.Difficulties[0])

	// Première case vide.
	r, c := m.firstEmpty()
	m.cursorR, m.cursorC = r, c
	want := m.solution[r][c]

	nm, _ := m.Update(runeKey(rune('0' + want)))
	m = nm.(Model)
	if m.board[r][c] != want {
		t.Fatalf("le chiffre %d n'a pas été placé (obtenu %d)", want, m.board[r][c])
	}
	if m.screen == scrWin {
		t.Fatal("victoire prématurée")
	}

	// Annuler restaure l'état.
	nm, _ = m.Update(runeKey('u'))
	m = nm.(Model)
	if m.board[r][c] != 0 {
		t.Fatal("l'annulation n'a pas rétabli la case vide")
	}

	// Remplir toute la grille => victoire.
	m.board = m.solution
	m.recompute()
	m.checkWin()
	if m.screen != scrWin {
		t.Fatal("la victoire n'a pas été détectée")
	}
	if strings.TrimSpace(m.View()) == "" {
		t.Fatal("l'écran de victoire est vide")
	}
}

func TestNotesMode(t *testing.T) {
	m := testModel(120, 50)
	m = m.start(sudoku.Difficulties[0])
	r, c := m.firstEmpty()
	m.cursorR, m.cursorC = r, c

	// Activer les notes puis saisir 3 et 7.
	nm, _ := m.Update(runeKey('n'))
	m = nm.(Model)
	if !m.noteMode {
		t.Fatal("le mode notes ne s'est pas activé")
	}
	nm, _ = m.Update(runeKey('3'))
	m = nm.(Model)
	nm, _ = m.Update(runeKey('7'))
	m = nm.(Model)
	if got := m.notes[r][c]; got != (1<<3)|(1<<7) {
		t.Fatalf("notes attendues 3+7, obtenu %b", got)
	}
	if m.board[r][c] != 0 {
		t.Fatal("les notes ne doivent pas remplir la case")
	}
}

func TestHelpOverlay(t *testing.T) {
	m := testModel(120, 50)
	m = m.start(sudoku.Difficulties[0])
	nm, _ := m.Update(runeKey('?'))
	m = nm.(Model)
	if !m.showHelp {
		t.Fatal("l'aide ne s'est pas ouverte")
	}
	if !strings.Contains(m.View(), "Aide") {
		t.Fatal("le texte d'aide est absent")
	}
}

func TestTooSmall(t *testing.T) {
	m := testModel(20, 8)
	if !strings.Contains(m.View(), "trop petit") {
		t.Fatal("le message « terminal trop petit » devrait s'afficher")
	}
}

// La grille haute doit basculer en compact quand la fenêtre est courte.
func TestResponsiveLayout(t *testing.T) {
	tall := testModel(120, 50)
	tall = tall.start(sudoku.Difficulties[0])
	if _, _, ch := tall.fitPlay(); ch != 3 {
		t.Fatalf("fenêtre haute : hauteur de case %d, attendu 3", ch)
	}
	short := testModel(120, 24)
	short = short.start(sudoku.Difficulties[0])
	lines, _, ch := short.fitPlay()
	if ch != 1 {
		t.Fatalf("fenêtre courte : hauteur de case %d, attendu 1", ch)
	}
	if len(lines) > short.height {
		t.Fatalf("la mise en page compacte déborde : %d lignes pour %d", len(lines), short.height)
	}
}

// La grille doit paraître carrée à l'écran : largeur ≈ aspect x hauteur.
// C'est le point qui manquait (grille 31x31 rendue ~2x trop haute).
func TestGridLooksSquare(t *testing.T) {
	for _, a := range []float64{1.6, 2.0, 2.2, 2.6} {
		for _, h := range []int{24, 40, 60} {
			m := testModel(200, h)
			m.aspect = a
			m = m.start(sudoku.Difficulties[0])
			_, cw, ch := m.fitPlay()
			mm := m
			mm.cw, mm.ch = cw, ch
			visualW := float64(mm.gridW())
			visualH := float64(mm.gridH()) * a
			ratio := visualW / visualH
			if ratio < 0.70 || ratio > 1.35 {
				t.Errorf("aspect=%.1f h=%d : %dx%d (cw=%d ch=%d) -> ratio visuel %.2f",
					a, h, mm.gridW(), mm.gridH(), cw, ch, ratio)
			}
		}
	}
}
