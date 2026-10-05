SUDOKU-TUI — un sudoku pour le terminal
=======================================

Écrit en Go avec les bibliothèques Charm :
  - Bubble Tea (architecture Elm : Model / Update / View)
  - Lip Gloss (styles, couleurs, mise en page)

--------------------------------------------------------------------
ÉTAT ACTUEL   (mis à jour le samedi 3 octobre 2026)
--------------------------------------------------------------------

  Fini et utilisable. Rien en chantier, rien de cassé.

    - Jeu complet : génération de grilles à solution UNIQUE, 4 niveaux,
      saisie, notes, annuler, indice, pause, minuteur, aide, victoire.
    - Mise en page adaptative (grille haute / compacte / trop petit) et
      grille rendue carrée à l'écran.
    - `go test ./...` passe : logique (sudoku/) + interface (ui/).
    - Binaire installé et fonctionnel (~/.local/bin/sudoku-tui).

  Ce qui n'est PAS fait (voir « À FAIRE / IDÉES » en bas) :
    - option « validation immédiate » (aide pour les jeunes) ;
    - table des scores (nom, date, durée, niveau) ;
    - changement de difficulté SANS quitter la partie en cours
      (aujourd'hui la difficulté se choisit au menu de départ).

  Prochaine étape naturelle : la table des scores, puis la validation
  immédiate.

--------------------------------------------------------------------
LIEUX
--------------------------------------------------------------------

  Sources ........... le présent dépôt
  Binaire installé .. ~/.local/bin/sudoku-tui      (dans le PATH)
  Compilation ....... bin/sudoku-tui               (ignoré par Git)
  Dépôt Git ......... le présent dépôt             (branche master)
  Dépôt distant ..... https://github.com/Gatshub/sudoku-tui
  Données de jeu .... aucune pour l'instant
                      (la table des scores ira dans
                       ~/.local/share/sudoku-tui/scores.json)
  Documentation ..... README.txt (ce fichier)

--------------------------------------------------------------------
LANCER
--------------------------------------------------------------------

    sudoku-tui

Au premier écran, choisir la difficulté (↑ ↓ puis Entrée, ou taper 1-4).

--------------------------------------------------------------------
NIVEAUX
--------------------------------------------------------------------

    Facile      36 indices      (≈  2 ms de génération)
    Moyen       30 indices      (≈  4 ms)
    Difficile   26 indices      (≈ 15 ms)
    Expert      23 indices      (≈ 40 ms)

  Moins il y a d'indices, plus il faut de déductions. Toutes les grilles ont
  une solution unique, garantie à la construction.

--------------------------------------------------------------------
TOUCHES
--------------------------------------------------------------------

    ↑ ↓ ← →  ou  h j k l     déplacer le curseur
    1 … 9                    saisir un chiffre (ou basculer une note en mode notes)
    0 / Retour arrière       effacer la case
    n  ou  Espace            activer / désactiver le mode notes
    u                        annuler le dernier coup
    H                        révéler la solution de la case (compté comme indice)
    r                        réinitialiser la grille en cours
    p                        mettre en pause (masque la grille)
    ?                        afficher / fermer l'aide
    q                        quitter
    souris                   cliquer sur une case pour la sélectionner

--------------------------------------------------------------------
CE QUI EST AFFICHÉ
--------------------------------------------------------------------

  - Grille encadrée de traits épais tous les 3x3 (blocs).
  - Curseur : fond bleu. Ligne / colonne / bloc du curseur : fond légèrement
    éclairci. Toutes les cases portant le même chiffre que le curseur : surlignées.
  - Chiffres donnés par l'énigme : blanc, gras.
  - Chiffres saisis : bleu clair. En cas de doublon (ligne/colonne/bloc) : rouge.
  - Ligne « Reste » : combien de fois chaque chiffre reste à placer (✓ = complet).
  - Minuteur, compteur d'erreurs et d'indices dans le sous-titre.

  GRILLE CARRÉE : un caractère de terminal est environ 2 fois plus haut que
  large. La largeur d'une case est donc calculée pour que la grille paraisse
  carrée à l'écran. Le rapport réel est déduit de la taille en pixels de la
  fenêtre quand le terminal la fournit ; sinon on suppose 2:1.

  Mise en page adaptative :
    - fenêtre haute  : cases de 3 lignes, notes en mini-grille 3x3  (~58 x 31)
    - fenêtre courte : cases sur 1 ligne, notes réduites à un point (~22 x 13)
    - fenêtre trop petite : message « Terminal trop petit »

--------------------------------------------------------------------
FONCTIONNEMENT INTERNE
--------------------------------------------------------------------

  sudoku/sudoku.go   Logique pure, sans dépendance UI :
                       - Board 9x9 (0 = case vide)
                       - solveur par backtracking avec heuristique MRV
                       - Conflicts() : cases en doublon
                       - Generate() : grille à solution UNIQUE, construite par
                         retrait symétrique (rotation 180°) de cases ; une passe
                         asymétrique de secours approche le nombre d'indices visé
  sudoku/*_test.go   Tests : solveur, conflits, unicité, cohérence des indices,
                     nombre d'indices réellement atteint par niveau
  ui/styles.go       Palette (Tokyo Night, adaptative clair/sombre) + styles
  ui/model.go        État, clavier, souris, minuteur, annuler, notes, pause,
                     et détection du rapport hauteur/largeur du terminal
  ui/view.go         Rendu : grille, en-tête, statut, menu, aide, victoire
  ui/*_test.go       Tests : alignement de la grille, grille carrée, pas de
                     débordement, déplacement du curseur, clic → case, notes,
                     victoire, mise en page adaptative

  Le générateur garantit à chaque étape l'unicité de la solution : la grille
  proposée n'a jamais qu'une seule solution, quel que soit le niveau.

--------------------------------------------------------------------
REBUILD / TESTS
--------------------------------------------------------------------

  Go est fourni par mise (go 1.27.1). Si « go » n'est pas trouvé :

      mise use -g go@1.27.1

  Ensuite :

      cd sudoku-tui
      make build      # compile -> bin/sudoku-tui
      make test       # lance tous les tests
      make run        # compile et lance
      make install    # installe dans ~/.local/bin
      make clean

  Pour voir le détail des niveaux :

      go test ./sudoku/ -run TestAchievedClueCounts -v

--------------------------------------------------------------------
DÉSINSTALLER
--------------------------------------------------------------------

      rm ~/.local/bin/sudoku-tui
      rm -rf sudoku-tui               # supprime aussi les sources

--------------------------------------------------------------------
À FAIRE / IDÉES
--------------------------------------------------------------------

  [ ] Option « validation immédiate » (aide pour les jeunes)
      But : qu'on sache TOUT DE SUITE qu'un chiffre placé n'est pas le bon,
      sans avoir à attendre qu'il crée un doublon visible.

      État actuel : on signale déjà les CONFLITS (deux fois le même chiffre
      sur une ligne / colonne / bloc), en rouge. Mais un chiffre faux qui ne
      duplique rien tout de suite passe inaperçu.

      Piste d'implémentation (le plus gros est déjà là) :
        - la solution complète est déjà conservée dans Model.solution ;
        - ajouter un champ  strictMode bool  dans Model ;
        - dans place(), si strictMode && valeur != solution[r][c] :
            * marquer la case (texte rouge, ou fond rouge léger) ;
            * message de statut « ✗ Ce n'est pas le bon chiffre » ;
          sinon comportement actuel.
        - la coloration se greffe sur cellColors() (même canal que les
          conflits, via le tableau m.conflicts ou un nouveau tableau) ;
        - l'annulation (u) fonctionne déjà : rien à ajouter.

      Reste à décider :
        - comment l'activer : choix au menu (comme la difficulté) et/ou
          touche en jeu (ex. « v »), avec un indicateur dans le sous-titre ;
        - la couleur : rouge franc (comme les conflits) ou plus doux pour ne
          pas décourager un enfant ;
        - faut-il la désactiver par défaut ? (probablement oui, c'est une aide)

  [ ] Table des scores (tableau d'honneur)
      But : à la fin d'une partie, garder une trace du score, et demander au
      joueur s'il veut le conserver. Inspiré de Nudoku.

      Colonnes prévues : nom du joueur, date, durée de la partie, niveau.

      Piste d'implémentation (le plus gros est déjà là) :
        - le minuteur est déjà dans Model.elapsed et le niveau dans
          Model.difficulty : tout ce qu'il faut pour remplir une ligne ;
        - nouveau paquet  scores/scores.go  (logique pure, sans UI) :
            * type Entry struct { Name string; Date time.Time;
              Duration time.Duration; Difficulty sudoku.Difficulty }
            * Load() / Save() sur un fichier de scores, format CSV ou JSON
              (lisible à la main, facile à tester) ;
            * Add(e Entry) qui insère et renvoie le classement trié ;
            * chemin de données : ~/.local/share/sudoku-tui/scores.json
              (respecter XDG_DATA_HOME), créé au besoin.
        - à l'écran de victoire (ui/view.go) : demander « Garder ce score ?
          (o/n) », puis, si oui, un petit champ de texte (bubbles/textinput)
          pour le nom — prérempli avec le dernier nom utilisé ;
        - touche pour consulter la table depuis le menu (ex. « s »).

      Reste à décider :
        - la table est-elle globale ou séparée par niveau ?
        - afficher toutes les lignes, ou seulement le top 10 + la partie qui
          vient d'être jouée ?
        - que faire si le joueur a utilisé des indices ou révélé la solution :
          garder quand même (avec une colonne « aide ») ou refuser le score ?

  [ ] (plus tard) Grille hexagonale : moteur de jeu de logique à cases
      hexagonales (coordonnées axiales, 6 voisins, rendu en quinconce).