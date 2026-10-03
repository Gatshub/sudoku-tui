package sudoku

import (
	"math/rand"
	"testing"
)

// Vérifie que le nombre d'indices réellement atteint est proche de la cible,
// et documente la difficulté effective de chaque niveau.
func TestAchievedClueCounts(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	const n = 20
	for _, d := range Difficulties {
		total, min, max := 0, 99, 0
		for i := 0; i < n; i++ {
			p, _ := Generate(d, rng)
			c := p.ClueCount()
			total += c
			if c < min {
				min = c
			}
			if c > max {
				max = c
			}
		}
		avg := float64(total) / float64(n)
		t.Logf("%-10s cible=%2d  obtenu: min=%2d  moy=%4.1f  max=%2d", d.Name, d.Clues, min, avg, max)
		if max > d.Clues+6 {
			t.Errorf("%s : la cible n'est pas respectée (max %d pour cible %d)", d.Name, max, d.Clues)
		}
	}
}

func BenchmarkGenerate(b *testing.B) {
	for _, d := range Difficulties {
		b.Run(d.Name, func(b *testing.B) {
			rng := rand.New(rand.NewSource(1))
			for i := 0; i < b.N; i++ {
				Generate(d, rng)
			}
		})
	}
}
