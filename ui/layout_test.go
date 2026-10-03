package ui

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"sudoku-tui/sudoku"
)

// Aucune ligne de l'affichage ne doit dépasser la largeur du terminal, à toutes
// les tailles raisonnables (c'est ce qui provoquait une aide tronquée).
func TestNoHorizontalOverflow(t *testing.T) {
	for _, w := range []int{36, 40, 50, 60, 70, 80, 100, 120, 160, 200} {
		for _, h := range []int{20, 24, 30, 38, 50, 60} {
			m := testModel(w, h)
			m = m.start(sudoku.Difficulties[0])
			for i, l := range strings.Split(m.View(), "\n") {
				if got := lipgloss.Width(l); got > w {
					t.Fatalf("w=%d h=%d : ligne %d fait %d colonnes", w, h, i, got)
				}
			}
		}
	}
}

// Le menu, l'aide et l'écran de victoire doivent eux aussi tenir en largeur.
func TestOtherScreensFit(t *testing.T) {
	for _, w := range []int{60, 80, 100, 140} {
		for _, h := range []int{24, 40, 50} {
			menu := testModel(w, h)
			for i, l := range strings.Split(menu.View(), "\n") {
				if lipgloss.Width(l) > w {
					t.Fatalf("menu w=%d : ligne %d trop large", w, i)
				}
			}

			help := testModel(w, h)
			help = help.start(sudoku.Difficulties[0])
			help.showHelp = true
			for i, l := range strings.Split(help.View(), "\n") {
				if lipgloss.Width(l) > w {
					t.Fatalf("aide w=%d : ligne %d trop large", w, i)
				}
			}

			win := testModel(w, h)
			win = win.start(sudoku.Difficulties[0])
			win.board = win.solution
			win.recompute()
			win.checkWin()
			for i, l := range strings.Split(win.View(), "\n") {
				if lipgloss.Width(l) > w {
					t.Fatalf("victoire w=%d : ligne %d trop large", w, i)
				}
			}
		}
	}
}

// En fenêtre courte, la grille doit rester complète et alignée en mode compact.
func TestCompactGridComplete(t *testing.T) {
	m := testModel(100, 24)
	m = m.start(sudoku.Difficulties[0])
	lines, _, ch := m.fitPlay()
	if ch != 1 {
		t.Fatalf("hauteur de case %d, attendu 1", ch)
	}
	// 9 lignes de cases + 2 séparateurs de blocs + 2 bords = 13
	var grid []string
	for _, l := range lines {
		if strings.Contains(l, bHoriz) || strings.Contains(l, bVert) {
			grid = append(grid, l)
		}
	}
	if len(grid) != 13 {
		t.Fatalf("grille compacte : %d lignes, attendu 13", len(grid))
	}
}

// L'écran de victoire doit réutiliser la grille carrée, pas retomber sur la
// largeur par défaut (c'était le cas avant : grille 31 colonnes au lieu de 58).
func TestWinGridUsesChosenLayout(t *testing.T) {
	for _, a := range []float64{1.6, 2.0, 2.6} {
		for _, h := range []int{24, 50, 60} {
			m := testModel(200, h)
			m.aspect = a
			m = m.start(sudoku.Difficulties[0])
			m.board = m.solution
			m.recompute()
			m.checkWin()

			cw, ch := m.chooseLayout(func(mm Model) []string { return mm.buildWin() })
			want := 9*cw + 4

			mm := m
			mm.cw, mm.ch = cw, ch
			found := false
			for _, l := range mm.buildWin() {
				if strings.Contains(l, bTopLeft) {
					if got := lipgloss.Width(l); got != want {
						t.Fatalf("aspect=%.1f h=%d : bordure de victoire %d colonnes, attendu %d",
							a, h, got, want)
					}
					found = true
				}
			}
			if !found {
				t.Fatalf("aspect=%.1f h=%d : aucune bordure dans l'écran de victoire", a, h)
			}
		}
	}
}

// Sur un terminal plus étroit que la largeur idéale, on réduit la largeur de
// case plutôt que de basculer d'un coup en grille compacte ; la grille reste
// plausible et ne dépasse jamais.
func TestNarrowTerminalDegradesGently(t *testing.T) {
	for _, w := range []int{40, 50, 58, 64, 70, 80} {
		m := testModel(w, 50)
		m.aspect = 2.0
		m = m.start(sudoku.Difficulties[0])
		_, cw, ch := m.fitPlay()
		mm := m
		mm.cw, mm.ch = cw, ch
		if mm.gridW() > w {
			t.Fatalf("w=%d : grille de %d colonnes, dépasse la fenêtre", w, mm.gridW())
		}
		ratio := float64(mm.gridW()) / (2.0 * float64(mm.gridH()))
		if ratio < 0.70 || ratio > 1.35 {
			t.Errorf("w=%d : ratio visuel %.2f (cw=%d ch=%d)", w, ratio, cw, ch)
		}
	}
}

// Aperçu textuel des écrans, pour vérifier la mise en page à l'œil sans
// terminal interactif. Hors TTY, Lip Gloss n'émet aucun code couleur : la
// sortie est du texte brut, directement lisible.
//
//	SUDOKU_PREVIEW=1 go test ./ui/ -run TestPreviewLayouts -v
func TestPreviewLayouts(t *testing.T) {
	if os.Getenv("SUDOKU_PREVIEW") == "" {
		t.Skip("définir SUDOKU_PREVIEW=1 pour afficher les écrans")
	}
	for _, tc := range []struct {
		name   string
		w, h   int
		aspect float64
	}{
		{"menu 120x50", 120, 50, 2.0},
		{"jeu 120x50 (aspect 2:1)", 120, 50, 2.0},
		{"jeu 120x50 (aspect 2.4:1)", 120, 50, 2.4},
		{"jeu 100x24 compact", 100, 24, 2.0},
	} {
		m := testModel(tc.w, tc.h)
		m.aspect = tc.aspect
		if strings.HasPrefix(tc.name, "jeu") {
			m = m.start(sudoku.Difficulties[0])
		}
		fmt.Printf("\n===== %s =====\n%s\n", tc.name, m.View())
	}

	// Écran de victoire, avec les notes du curseur.
	m := testModel(120, 50)
	m.aspect = 2.0
	m = m.start(sudoku.Difficulties[0])
	r, c := m.firstEmpty()
	m.cursorR, m.cursorC = r, c
	m.notes[r][c] = 1<<1 | 1<<3 | 1<<5 | 1<<7 | 1<<9
	m.board = m.solution
	m.recompute()
	m.checkWin()
	fmt.Printf("\n===== victoire 120x50 =====\n%s\n", m.View())
}
