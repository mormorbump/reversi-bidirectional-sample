package game

import "fmt"

// Board 盤面を8×8のセルとそれを囲む壁で表現する。そのため10×10の二次元配列となる(壁は上下左右で1列ずつなので、xとyは2ずつ引いて8×8)
type Board struct {
	// セルを定義。
	Cells [][]Character
}

const (
	cellNum          = 10
	wallThresholdNum = 9
)

// NewBoard 盤面を作成。壁を作成することで、セルを調べる際に壁かどうかを確認するだけで範囲外かどうかを判定する条件文を省略できる。
func NewBoard() *Board {
	// 8x8のセル+壁で、10x10の盤面を二次元配列で作成
	b := &Board{
		Cells: make([][]Character, cellNum),
	}
	for i := 0; i < cellNum; i++ {
		b.Cells[i] = make([]Character, cellNum)
	}

	// 盤面の端に壁を設置。
	// 左の壁。上下(0,0)(0,9)も含むので0 <= i < 10
	for i := 0; i < cellNum; i++ {
		b.Cells[0][i] = Wall
	}

	// 上下の壁。左の壁作成時に(0,0)(0,9)は作ってあるので1 <= i < 9
	for i := 1; i < wallThresholdNum; i++ {
		b.Cells[i][0] = Wall
		b.Cells[i][wallThresholdNum] = Wall
	}

	// 右の壁。左上下の壁作成時に(0,9), (9,9)は作ってあるので1 <= i < 9
	for i := 0; i < wallThresholdNum; i++ {
		b.Cells[wallThresholdNum][i] = Wall
	}

	// 初期石
	b.Cells[4][4] = White
	b.Cells[5][5] = White
	b.Cells[5][4] = Black
	b.Cells[4][5] = Black

	return b
}

func (b *Board) PutStone(x int32, y int32, c Character) error {
	// セルに石を置けるかチェック
	if !b.CanPutStone(x, y, c) {
		return fmt.Errorf("can not put stone x=%v, y=%v color=%v", x, y, CharacterToStr(c))
	}

	b.Cells[x][y] = c

	// 置いた石の縦/横/斜めの各方向でひっくり返すことのできる石を全てひっくり返す
	// 各方向 <=> (1,1), (1,-1), (-1,1), (-1,-1)と言うこと
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			// 自分自身はスキップ
			if dx == 0 && dy == 0 {
				continue
			}
			b.TurnStonesByDirection(x, y, c, int32(dx), int32(dy))
		}
	}
	return nil
}

func (b *Board) CanPutStone(x int32, y int32, c Character) bool {
	// すでに石が置いてあったらng
	if b.Cells[x][y] != Empty {
		return false
	}

	// 置いた石の縦/横/斜めの各方向をチェック
	// 各方向 <=> (1,1), (1, 0), (1,-1), (0, 1), (-1,1), (-1,-1)と言うこと
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			// (0,0)は自分自身なのでスキップ
			if dx == 0 && dy == 0 {
				continue
			}
			// ひっくり返すことのできる石が一つでもあれば石をおけるのでtrue
			if b.CountTurnableStonesByDirection(x, y, c, int32(dx), int32(dy)) > 0 {
				return true
			}
		}
	}

	// 置けると来なかったのでfalse
	return false
}

// CountTurnableStonesByDirection あるセルに石を置いた場合、ある方向にひっくり返すことのできる石がいくつあるかをカウント
// 方向を固定して走査したときに、相手の石が連続して並んでいる&&終端に自分の意思がある場合のみ条件とする
func (b *Board) CountTurnableStonesByDirection(x int32, y int32, c Character, dx int32, dy int32) int {
	cnt := 0

	// 置いた石を基準として、1方向ずらしたセルの位置。
	nx := x + dx
	ny := y + dy

	for {
		nc := b.Cells[nx][ny]

		// 壁か自分の石であればループを終了
		if nc != OpponentCharacter(c) {
			break
		}

		// 相手の石なので数え上げ
		cnt++

		// さらにひっくり返すかチェックするためdx, dy分追加
		nx += dx
		ny += dy
	}

	// その方向にある相手の石の数がゼロより大きく、かつそのさきに自分の石がある場合は数を返す
	// このチェックが必要な場合は、その方向に相手の石しか存在しない場合はひっくり返せないから(自分の石で挟み込む必要あり)
	if cnt > 0 && b.Cells[nx][ny] == c {
		return cnt
	}

	// それ以外の場合はゼロ
	return 0
}

// TurnStonesByDirection ある方向の石をひっくり返す。(ひっくり返して良い場合だけよぶ)
func (b *Board) TurnStonesByDirection(x int32, y int32, c Character, dx int32, dy int32) {
	nx := x + dx
	ny := y + dy

	for {
		nc := b.Cells[nx][ny]
		if nc != OpponentCharacter(c) {
			break
		}

		b.Cells[nx][ny] = c
		nx += dx
		ny += dy
	}
}

// AvailableCellCount 盤面内で、「ある色の石」をおけるセルの数を数える
func (b *Board) AvailableCellCount(c Character) int {
	cnt := 0
	// iはwall以外の盤面全てを探索する
	for i := 1; i < wallThresholdNum; i++ {
		for j := 1; j < wallThresholdNum; j++ {
			if b.CanPutStone(int32(i), int32(j), c) {
				cnt++
			}
		}
	}
	return cnt
}

// Score 盤面内に置かれている石の数
func (b *Board) Score(c Character) int {
	cnt := 0
	for i := 1; i < wallThresholdNum; i++ {
		for j := 1; j < wallThresholdNum; j++ {
			// 自分のCharacterじゃなかったらskip
			if b.Cells[i][j] != c {
				continue
			}
			cnt++
		}
	}

	return cnt
}

// Rest 盤面ないで石が置かれていないセルの数
func (b *Board) Rest() int {
	cnt := 0
	for i := 1; i < wallThresholdNum; i++ {
		for j := 1; j < wallThresholdNum; j++ {
			if b.Cells[i][j] == Empty {
				cnt++
			}
		}
	}
	return cnt
}
