package game

import (
	"math/rand"
)

type Drawable interface {
	GetRatio() int
}

type Reward struct {
	CardID string
	Ratio  int
}

func (e *Reward) GetRatio() int {
	return e.Ratio
}

// Rewards 1. 各アイテムの提供割合データを用意
var Rewards = []*Reward{
	{CardID: "Sレアカード", Ratio: 1},
	{CardID: "レアカード", Ratio: 3},
	{CardID: "ノーマルカード", Ratio: 6},
}

// ジェネリクスを使わない書き方
func draw(drawables []Drawable) Drawable {
	// 2. 提供割合の合計値を計算
	var total int
	for _, d := range drawables {
		total += d.GetRatio()
	}
	// 3. 合計値の範囲で乱数を生成
	random := rand.Intn(total)

	// 4. 乱数を元に抽選結果を決定
	var temp int
	for _, d := range drawables {
		// ratioを足していくことで次のcardを抽選することができる。
		temp += d.GetRatio()
		if temp > random {
			return d
		}
	}
	return nil
}

// Draw ジェネリクスを使わないと、呼び出し元が煩雑
// 関数の引数に合うようにinterface型の配列に詰めた後、さらに元に戻す必要がある。
func Draw() string {
	// Rewardsを[]Drawableに詰め替えて、draw関数を実行
	drawables := make([]Drawable, 0, len(Rewards))
	for _, e := range Rewards {
		drawables = append(drawables, e)
	}
	// draw関数の戻り値(Drawable)を*Rewardにキャストして値を取得
	cardId := draw(drawables).(*Reward).CardID
	return "抽選されたカード: " + cardId
}

// drawGenerics Drawableインターフェース型の型パラメータTを定義し、引数、戻り値もTに変更して汎用化
func drawGenerics[T Drawable](drawables []T) T {
	// 2. 提供割合の合計値を計算
	var total int
	for _, d := range drawables {
		total += d.GetRatio()
	}

	// 3. 合計値の範囲で乱数を生成
	random := rand.Intn(total)

	// 4. 乱数を元に抽選結果を決定
	var temp int
	for _, d := range drawables {
		// ratioを足していくことで次のcardを抽選することができる。
		temp += d.GetRatio()
		if temp > random {
			return d
		}
	}

	// return nilだとコンパイルエラーになる
	var ret T
	return ret
}

func DrawGenerics() string {
	// 呼び出す側でRewards直接渡せる。Drawableへのキャストが暗黙的に行われる
	return "抽選されたカード: " + drawGenerics(Rewards).CardID
}
