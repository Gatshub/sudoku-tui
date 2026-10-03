package sudoku

import (
	"math/rand"
	"testing"
)

func TestSolveKnown(t *testing.T) {
	// Énigme classique (Wikipedia) et sa solution.
	raw := [9][9]int{
		{5, 3, 0, 0, 7, 0, 0, 0, 0},
		{6, 0, 0, 1, 9, 5, 0, 0, 0},
		{0, 9, 8, 0, 0, 0, 0, 6, 0},
		{8, 0, 0, 0, 6, 0, 0, 0, 3},
		{4, 0, 0, 8, 0, 3, 0, 0, 1},
		{7, 0, 0, 0, 2, 0, 0, 0, 6},
		{0, 6, 0, 0, 0, 0, 2, 8, 0},
		{0, 0, 0, 4, 1, 9, 0, 0, 5},
		{0, 0, 0, 0, 8, 0, 0, 7, 9},
	}
	var b Board = raw
	sol, ok := b.Solve()
	if !ok {
		t.Fatal("la grille devrait être résoluble")
	}
	if !sol.IsSolved() {
		t.Fatal("la solution renvoyée n'est pas valide")
	}
	// Les cases données doivent être conservées.
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if raw[r][c] != 0 && sol[r][c] != raw[r][c] {
				t.Fatalf("case donnée modifiée en (%d,%d)", r, c)
			}
		}
	}
}

func TestConflicts(t *testing.T) {
	var b Board
	b[0][0], b[0][5] = 4, 4 // même ligne
	b[8][0] = 4             // même colonne
	b[1][1] = 4             // même bloc
	cf := b.Conflicts()
	for _, cell := range [][2]int{{0, 0}, {0, 5}, {8, 0}, {1, 1}} {
		if !cf[cell[0]][cell[1]] {
			t.Errorf("la case (%d,%d) devrait être en conflit", cell[0], cell[1])
		}
	}
	if cf[4][4] {
		t.Error("une case isolée ne doit pas être en conflit")
	}
}

func TestGenerateUniqueAndSolvable(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for _, d := range Difficulties {
		for i := 0; i < 8; i++ {
			puzzle, solution := Generate(d, rng)
			if !solution.IsSolved() {
				t.Fatalf("[%s] la solution générée est invalide", d.Name)
			}
			if !puzzle.HasUniqueSolution() {
				t.Fatalf("[%s] l'énigme n'a pas une solution unique", d.Name)
			}
			// Les indices de l'énigme doivent correspondre à la solution.
			for r := 0; r < 9; r++ {
				for c := 0; c < 9; c++ {
					if puzzle[r][c] != 0 && puzzle[r][c] != solution[r][c] {
						t.Fatalf("[%s] indice incohérent en (%d,%d)", d.Name, r, c)
					}
				}
			}
			// Le solveur doit retrouver exactement la solution.
			solved, ok := puzzle.Solve()
			if !ok || solved != solution {
				t.Fatalf("[%s] le solveur ne retrouve pas la solution unique", d.Name)
			}
			if got := puzzle.ClueCount(); got < 17 {
				t.Fatalf("[%s] trop peu d'indices : %d", d.Name, got)
			}
		}
	}
}

func TestIsSolvedEmpty(t *testing.T) {
	var b Board
	if b.IsSolved() {
		t.Fatal("une grille vide n'est pas résolue")
	}
}

func TestValidBounds(t *testing.T) {
	var b Board
	if b.Valid(0, 0, 0) || b.Valid(0, 0, 10) {
		t.Fatal("les valeurs hors 1..9 doivent être refusées")
	}
	if !b.Valid(0, 0, 1) {
		t.Fatal("1 doit être valide sur grille vide")
	}
}
