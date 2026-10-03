// Package sudoku implémente la logique d'un Sudoku 9x9 : résolution,
// génération de grilles à solution unique et détection de conflits.
// Le paquet est volontairement pur (aucune dépendance UI).
package sudoku

import (
	"math/rand"
)

// Board est une grille 9x9. 0 signifie « case vide ».
type Board [9][9]int

// Difficulty décrit un niveau de difficulté par son nombre de cases révélées.
type Difficulty struct {
	Name  string
	Clues int // nombre de cases remplies dans la grille de départ
}

// Difficulties liste les niveaux proposés, du plus simple au plus dur.
// Le nombre d'indices (cases révélées) est le principal levier de difficulté :
// moins il y a d'indices, plus il faut de déductions pour avancer.
var Difficulties = []Difficulty{
	{Name: "Facile", Clues: 36},
	{Name: "Moyen", Clues: 30},
	{Name: "Difficile", Clues: 26},
	{Name: "Expert", Clues: 23},
}

// At renvoie la valeur d'une case (0 = vide).
func (b Board) At(r, c int) int { return b[r][c] }

// Set fixe la valeur d'une case.
func (b *Board) Set(r, c, v int) { b[r][c] = v }

// Valid indique si poser v en (r,c) respecte les règles du Sudoku.
func (b Board) Valid(r, c, v int) bool {
	if v < 1 || v > 9 {
		return false
	}
	for i := 0; i < 9; i++ {
		if i != c && b[r][i] == v {
			return false
		}
		if i != r && b[i][c] == v {
			return false
		}
	}
	br, bc := r/3*3, c/3*3
	for i := br; i < br+3; i++ {
		for j := bc; j < bc+3; j++ {
			if (i != r || j != c) && b[i][j] == v {
				return false
			}
		}
	}
	return true
}

// IsSolved indique si la grille est complète et valide.
func (b Board) IsSolved() bool {
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if b[r][c] == 0 || !b.Valid(r, c, b[r][c]) {
				return false
			}
		}
	}
	return true
}

// Conflicts marque toutes les cases qui dupliquent une valeur sur leur
// ligne, colonne ou bloc 3x3. Sert à colorer les erreurs en rouge.
func (b Board) Conflicts() (conflict [9][9]bool) {
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			v := b[r][c]
			if v == 0 {
				continue
			}
			for i := 0; i < 9; i++ {
				if i != c && b[r][i] == v {
					conflict[r][c], conflict[r][i] = true, true
				}
				if i != r && b[i][c] == v {
					conflict[r][c], conflict[i][c] = true, true
				}
			}
			br, bc := r/3*3, c/3*3
			for i := br; i < br+3; i++ {
				for j := bc; j < bc+3; j++ {
					if (i != r || j != c) && b[i][j] == v {
						conflict[r][c], conflict[i][j] = true, true
					}
				}
			}
		}
	}
	return
}

// candidates compte les valeurs légales pour une case vide.
func (b Board) candidates(r, c int) (mask uint16, n int) {
	for v := 1; v <= 9; v++ {
		if b.Valid(r, c, v) {
			mask |= 1 << uint(v)
			n++
		}
	}
	return
}

// bestCell choisit la case vide au moins de candidats (heuristique MRV),
// ce qui rend le backtracking très rapide.
func (b Board) bestCell() (int, int, bool) {
	bestR, bestC, bestN := -1, -1, 10
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if b[r][c] != 0 {
				continue
			}
			_, n := b.candidates(r, c)
			if n < bestN {
				bestR, bestC, bestN = r, c, n
				if n <= 1 {
					return bestR, bestC, true
				}
			}
		}
	}
	if bestR == -1 {
		return -1, -1, false // plus de case vide : résolu
	}
	return bestR, bestC, true
}

// solve remplit la grille en place. Si rng != nil, l'ordre des valeurs est
// mélangé, ce qui permet de produire des solutions variées.
func (b *Board) solve(rng *rand.Rand) bool {
	r, c, ok := b.bestCell()
	if !ok {
		return true
	}
	order := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	if rng != nil {
		rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
	}
	for _, v := range order {
		if b.Valid(r, c, v) {
			b[r][c] = v
			if b.solve(rng) {
				return true
			}
			b[r][c] = 0
		}
	}
	return false
}

// Solve résout une copie de la grille et renvoie le résultat + un booléen.
func (b Board) Solve() (Board, bool) {
	out := b
	return out, out.solve(nil)
}

// countSolutions compte les solutions, en s'arrêtant dès que limit est atteint.
func (b *Board) countSolutions(limit int) int {
	r, c, ok := b.bestCell()
	if !ok {
		return 1
	}
	total := 0
	for v := 1; v <= 9; v++ {
		if b.Valid(r, c, v) {
			b[r][c] = v
			total += b.countSolutions(limit)
			b[r][c] = 0
			if total >= limit {
				return total
			}
		}
	}
	return total
}

// HasUniqueSolution indique si la grille admet exactement une solution.
func (b Board) HasUniqueSolution() bool {
	cp := b
	return cp.countSolutions(2) == 1
}

// Generate produit une grille (énigme) et sa solution. La grille est construite
// par retrait symétrique (rotation 180°) de cases, en garantissant à chaque
// étape l'unicité de la solution. Une passe finale asymétrique est tentée si le
// nombre de cases révélées visé n'est pas atteint.
func Generate(d Difficulty, rng *rand.Rand) (puzzle, solution Board) {
	if rng == nil {
		rng = rand.New(rand.NewSource(1))
	}

	// 1) Solution complète aléatoire.
	_ = solution.solve(rng)

	// 2) Retrait symétrique.
	puzzle = solution
	type pair struct{ r1, c1, r2, c2 int }
	var pairs []pair
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			r2, c2 := 8-r, 8-c
			if r*9+c <= r2*9+c2 {
				pairs = append(pairs, pair{r, c, r2, c2})
			}
		}
	}
	rng.Shuffle(len(pairs), func(i, j int) { pairs[i], pairs[j] = pairs[j], pairs[i] })

	clues := 81
	for _, p := range pairs {
		if clues <= d.Clues {
			break
		}
		a, b2 := puzzle[p.r1][p.c1], puzzle[p.r2][p.c2]
		if a == 0 && b2 == 0 {
			continue
		}
		puzzle[p.r1][p.c1], puzzle[p.r2][p.c2] = 0, 0
		if puzzle.HasUniqueSolution() {
			if a != 0 {
				clues--
			}
			if b2 != 0 {
				clues--
			}
		} else {
			puzzle[p.r1][p.c1], puzzle[p.r2][p.c2] = a, b2
		}
	}

	// 3) Passe asymétrique de secours pour approcher la cible.
	if clues > d.Clues {
		type cell struct{ r, c int }
		var cells []cell
		for r := 0; r < 9; r++ {
			for c := 0; c < 9; c++ {
				if puzzle[r][c] != 0 {
					cells = append(cells, cell{r, c})
				}
			}
		}
		rng.Shuffle(len(cells), func(i, j int) { cells[i], cells[j] = cells[j], cells[i] })
		for _, cl := range cells {
			if clues <= d.Clues {
				break
			}
			save := puzzle[cl.r][cl.c]
			puzzle[cl.r][cl.c] = 0
			if puzzle.HasUniqueSolution() {
				clues--
			} else {
				puzzle[cl.r][cl.c] = save
			}
		}
	}

	return puzzle, solution
}

// Clone renvoie une copie indépendante de la grille.
func (b Board) Clone() Board { return b }

// ClueCount compte les cases remplies.
func (b Board) ClueCount() int {
	n := 0
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if b[r][c] != 0 {
				n++
			}
		}
	}
	return n
}
