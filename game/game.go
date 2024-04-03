package game

import "fmt"

type Game struct {
	Board    *Board
	started  bool
	finished bool
	me       Character
}

func NewGame(me Character) *Game {
	return &Game{
		Board: NewBoard(),
		me:    me,
	}
}

// Move 手を打血、その後盤面を出力する
// 返り値として、ゲームが終了したかを返却
// TODO: Progressなどに命名変更するべき
func (g *Game) Move(x int32, y int32, c Character) (bool, error) {
	if g.finished {
		return true, nil
	}
	err := g.Board.PutStone(x, y, c)
	if err != nil {
		return false, err
	}
	// TODO: この引数いる？？
	g.Display()
	if g.IsGameOver() {
		fmt.Println("finished")
		g.finished = true
		return true, nil
	}
	return false, nil
}

// IsGameOver ゲームが終了したかを判定
// 黒と白双方における場所がなければ終了とする
func (g *Game) IsGameOver() bool {
	if g.Board.AvailableCellCount(Black) > 0 || g.Board.AvailableCellCount(White) > 0 {
		return false
	}
	return true
}

// Winner 勝者の色を返却。引き分けの場合はNone
func (g *Game) Winner() Character {
	black := g.Board.Score(Black)
	white := g.Board.Score(White)
	if black == white {
		return None
	} else if black > white {
		return Black
	}
	return White
}

// Display 盤面を出力
func (g *Game) Display() {
	fmt.Println("")
	if g.me != None {
		fmt.Printf("You: %v\n", CharacterToStr(g.me))
	}

	fmt.Print(" ｜ ")
	rs := []rune("ABCDEFGH")
	for i, r := range rs {
		fmt.Printf("%v", string(r))
		if i < len(rs)-1 {
			fmt.Print(" ｜ ")
		}
	}
	fmt.Print("\n")
	fmt.Println("ーーーーーーーーーーーーーー")

	for j := 1; j < wallThresholdNum; j++ {
		fmt.Printf("%d", j)
		fmt.Print(" ｜ ")
		for i := 1; i < wallThresholdNum; i++ {
			fmt.Print(CharacterToStr(g.Board.Cells[i][j]))
			fmt.Print(" ｜ ")
		}
		fmt.Print("\n")
	}

	fmt.Println("ーーーーーーーーーーーーーー")

	fmt.Printf("Score: BLACK=%d, WHITE=%d REST=%d\n",
		g.Board.Score(Black), g.Board.Score(White),
		g.Board.Rest(),
	)

	fmt.Print("\n")

}
