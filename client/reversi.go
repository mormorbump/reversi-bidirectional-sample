package client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"kazuki.matsumoto/reversi/build"
	"kazuki.matsumoto/reversi/game"
	"kazuki.matsumoto/reversi/gen/pb"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Reversi struct {
	sync.RWMutex
	started  bool
	finished bool
	isColor  game.Character //手番を表す
	me       *game.Player
	room     *game.Room
	game     *game.Game
}

func NewReversi() *Reversi {
	return &Reversi{
		isColor: game.Black,
	}
}

func (r *Reversi) Run() int {
	if err := r.run(); err != nil {
		fmt.Println(err)
		return 1
	}
	return 0
}

func (r *Reversi) run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		return errors.New("failed to connect to grpc server")
	}
	defer conn.Close()

	// マッチング問い合わせ
	err = r.matching(ctx, pb.NewMatchingServiceClient(conn))
	if err != nil {
		return err
	}

	// マッチングできたので盤面作成
	r.game = game.NewGame(r.me.Character)

	// 双方向ストリーミングでゲーム処理
	return r.play(ctx, pb.NewGameServiceClient(conn))
}

func (r *Reversi) matching(ctx context.Context, cli pb.MatchingServiceClient) error {
	// マッチングリクエスト
	stream, err := cli.JoinRoom(ctx, &pb.JoinRoomRequest{})
	if err != nil {
		return err
	}
	defer stream.CloseSend()

	fmt.Println("Requested matching...")

	// ストリーミングでレスポンスを受け取る
	for {
		resp, err := stream.Recv()
		if err != nil {
			return err
		}
		// マッチング成立
		if resp.GetStatus() == pb.JoinRoomResponse_MATCHED {
			r.room = build.Room(resp.GetRoom())
			r.me = build.Player(resp.GetMe())
			fmt.Printf("Matched room_id=%\n", resp.GetRoom().GetId())
			return nil
		} else if resp.GetStatus() == pb.JoinRoomResponse_WAITING {
			fmt.Println("Waiting matching...")
		}
	}
}

func (r *Reversi) play(ctx context.Context, cli pb.GameServiceClient) error {
	c, cancel := context.WithCancel(ctx)
	defer cancel()

	// 双方向ストリーミングを開始する
	stream, err := cli.Play(c)
	if err != nil {
		return err
	}
	defer stream.CloseSend()

	go func() {
		// 自分の手を送信
		err := r.send(c, stream)
		if err != nil {
			cancel()
		}
	}()

	// 相手からの手を受信
	err = r.receive(c, stream)
	if err != nil {
		cancel()
		return err
	}

	return nil
}

func (r *Reversi) reset() {
	r.Lock()
	defer r.Unlock()
	r.started = false
	r.finished = false
	r.isColor = game.Black
	r.me = nil
	r.room = nil
	r.game = nil
}

func (r *Reversi) send(ctx context.Context, stream pb.GameService_PlayClient) error {
	for {
		// sendを送る時、recv側のfinishedやstartedが変更しないようにする
		r.RLock()

		// receive側で終了されたので、send側も終了する
		if r.finished {
			// 特にWriteするのもないのでUnlock
			r.RUnlock()
			r.reset()
			return nil
			// 未開始なので、開始リクエストを送る
		} else if !r.started {
			err := stream.Send(&pb.PlayRequest{
				RoomId: r.room.ID,
				Player: build.PBPlayer(r.me),
				Action: &pb.PlayRequest_Start{
					Start: &pb.StartAction{},
				},
			})
			// isStartedになるので、相手にStartActionを送ってからUnlock
			r.RUnlock()
			if err != nil {
				return err
			}

			// 相手が開始するまで待機するためのfor文
			for {
				r.RLock()
				if r.started {
					// 開始をreceiveし、for文をbreakすることでstart
					r.RUnlock()
					fmt.Println("Ready go!")
					break
				}
				r.RUnlock()
				fmt.Println("Waiting until opponent player ready")
				time.Sleep(1 * time.Second)
			}
		} else {
			// else以下になったら対戦中
			r.RUnlock()

			// 自分の手番でない場合はスキップ
			if r.isColor != r.me.Character {
				continue
			}

			// 手の入力を待機
			fmt.Print("Input Your Move (ex. A-1):")
			stdin := bufio.NewScanner(os.Stdin)
			stdin.Scan()

			// 入力された手を解析
			text := stdin.Text()
			x, y, err := parseInput(text)
			if err != nil {
				fmt.Println(err)
				continue
			}

			// 手を打つ
			r.Lock()
			_, err = r.game.Move(x, y, r.me.Character)
			if err != nil {
				r.Unlock()
				fmt.Println(err)
				continue
			}

			// サーバーに手を送る処理
			go func() {
				err = stream.Send(&pb.PlayRequest{
					RoomId: r.room.ID,
					Player: build.PBPlayer(r.me),
					Action: &pb.PlayRequest_Move{
						Move: &pb.MoveAction{
							Move: &pb.Move{
								X: x,
								Y: y,
							},
						},
					},
				})
				if err != nil {
					r.Unlock()
					fmt.Println(err)
				}

				r.isColor = game.OpponentCharacter(r.me.Character)
				r.Unlock()
			}()
		}

		select {
		case <-ctx.Done():
			// キャンセルされたので終了
			return nil
		default:
		}
	}
}

func (r *Reversi) receive(ctx context.Context, stream pb.GameService_PlayClient) error {
	for {
		res, err := stream.Recv()
		if err != nil {
			return err
		}

		r.Lock()
		// 送られてきたresponseのeventからやることを分岐
		switch res.GetEvent().(type) {
		case *pb.PlayResponse_Waiting:
			// 開始待機中(なので処理せず)
		case *pb.PlayResponse_Ready:
			// 開始
			r.started = true
			r.game.Display()
		case *pb.PlayResponse_Move:
			// 手を打たれた
			character := build.Character(res.GetMove().GetPlayer().GetCharacter())
			if character != r.me.Character {
				move := res.GetMove().GetMove()
				// クライアント側のゲーム情報に反映
				_, err = r.game.Move(move.GetX(), move.GetY(), character)
				if err != nil {
					return err
				}
				// 相手の手番が終わったので自分の手番に変更
				// 送信側でも色を変えてるが、プロセスが別れている==メモリも別れているので、こちらも変更の必要がある。
				r.isColor = r.me.Character
				fmt.Print("Input Your Move (ex. A-1):")
			}
		case *pb.PlayResponse_Finished:
			r.finished = true

			// 勝敗表示
			winner := build.Character(res.GetFinished().Winner)
			fmt.Println("")
			if winner == game.None {
				fmt.Println("Draw!")
			} else if winner == r.me.Character {
				fmt.Println("You Win!")
				s := game.DrawGenerics() // 抽選
				fmt.Println(s)
			} else {
				fmt.Println("You Lose!")
			}
			r.Unlock()
			// ループ終了するのでreturn
			return nil
		}
		r.Unlock()

		select {
		case <-ctx.Done():
			// キャンセルされたので終了
			return nil
		default:
		}
	}
}

// `A-2`の形式で入力された手を(x, y)=(1, 2)の形式に変換する
func parseInput(txt string) (int32, int32, error) {
	ss := strings.Split(txt, "-")
	if len(ss) != 2 {
		return 0, 0, fmt.Errorf("入力が不正です。例:A-1")
	}

	xs := ss[0]                        // B
	xrs := []rune(strings.ToUpper(xs)) // xsを大文字にして、runeでunicodeにする。 B -> 66
	x := int32(xrs[0]-rune('A')) + 1   // Bのコードポイント(66)からAのコードポイント(65)をひき、1スタートなので2とする。

	if x < 1 || 8 < x {
		return 0, 0, fmt.Errorf("入力が不正です。例:A-1")
	}

	ys := ss[1]
	y, err := strconv.ParseInt(ys, 10, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("入力が不正です。例:A-1")
	}

	if y < 1 || 8 < y {
		return 0, 0, fmt.Errorf("入力が不正です。例:A-1")
	}

	return x, int32(y), nil
}
