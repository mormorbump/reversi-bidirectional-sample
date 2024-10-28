package handler

import (
	"fmt"
	"kazuki.matsumoto/reversi/build"
	"kazuki.matsumoto/reversi/game"
	"kazuki.matsumoto/reversi/gen/pb"
	"sync"
)

type GameHandler struct {
	pb.UnimplementedGameServiceServer
	sync.RWMutex
	games  map[int32]*game.Game                  // ゲーム情報(盤面など)を格納
	client map[int32][]pb.GameService_PlayServer // 状態変更時にクライアントにストリーミングを返すために格納
}

const RoomJoinNum = 2

func NewGameHandler() *GameHandler {
	return &GameHandler{
		games:  make(map[int32]*game.Game),
		client: make(map[int32][]pb.GameService_PlayServer),
	}
}

// Play エントリーポイント。streamの中のActionによって処理が振り分けられる
func (h *GameHandler) Play(stream pb.GameService_PlayServer) error {
	for {
		// クライアントからリクエストを受信したら、reqにリクエストが代入
		req, err := stream.Recv()
		if err != nil {
			return err
		}

		// TODO この辺りをメモリ or interceptorから受け取る
		roomID := req.GetRoomId()
		player := build.Player(req.GetPlayer())

		// oneofで複数の型のリクエストがくるので、switch文で処理
		// TODO: oneof使うとこういうことになるのであんまやりたくないね
		switch req.GetAction().(type) {
		case *pb.PlayRequest_Start:
			// ゲーム開始リクエスト
			err := h.start(stream, roomID)
			if err != nil {
				return err
			}
		case *pb.PlayRequest_Move:
			// 石を置いたときのリクエスト
			action := req.GetMove()
			x := action.GetMove().GetX()
			y := action.GetMove().GetY()
			err := h.move(roomID, x, y, player)
			if err != nil {
				return err
			}
		}
	}
}

func (h *GameHandler) start(stream pb.GameService_PlayServer, roomID int32) error {
	h.Lock()
	defer h.Unlock()

	// mutexでロックしたいので、読み込みを一回にするためにメモ化
	g := h.games[roomID]

	// ゲーム情報がなければ作成する
	if g == nil {
		g = game.NewGame(game.None) // gameのインスタンス生成
		h.games[roomID] = g
		h.client[roomID] = make([]pb.GameService_PlayServer, 0, RoomJoinNum) // 2人分のstreamを格納し、clientに状態変更の通知をする準備をする
	}

	// 自分のクライアントを格納
	h.client[roomID] = append(h.client[roomID], stream)

	if len(h.client[roomID]) == RoomJoinNum {
		// 二人揃ったので開始。参加者全員のclientにブロードキャスト
		for _, s := range h.client[roomID] {
			err := s.Send(&pb.PlayResponse{
				Event: &pb.PlayResponse_Ready{
					Ready: &pb.PlayResponse_ReadyEvent{},
				},
			})
			if err != nil {
				return err
			}
		}
		fmt.Printf("game has started room_id=%v\n", roomID)
	} else {
		//まだroomが全員揃ってないので、待機中であることをクライアントに通知
		err := stream.Send(&pb.PlayResponse{
			Event: &pb.PlayResponse_Waiting{
				Waiting: &pb.PlayResponse_WaitingEvent{},
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *GameHandler) move(roomID int32, x int32, y int32, p *game.Player) error {
	h.Lock()

	// TODO 終了した時、ここでクライアントやGame構造体のmapからGameの内容を削除する処理を入れる
	defer h.Unlock()
	// mutexでロックしたいので、読み込みを一回にするためにメモ化
	g := h.games[roomID]

	finished, err := g.Move(x, y, p.Character)
	if err != nil {
		return err
	}

	for _, s := range h.client[roomID] {
		// 手が打たれたことをクライアントに通知
		err := s.Send(&pb.PlayResponse{
			Event: &pb.PlayResponse_Move{
				Move: &pb.PlayResponse_MoveEvent{
					Player: build.PBPlayer(p),
					Move: &pb.Move{
						X: x,
						Y: y,
					},
					Board: build.PBBoard(g.Board),
				},
			},
		})
		if err != nil {
			return err
		}

		if finished {
			// ゲーム終了を通知
			err := s.Send(
				&pb.PlayResponse{
					Event: &pb.PlayResponse_Finished{
						Finished: &pb.PlayResponse_FinishedEvent{
							Winner: build.PBCharacter(g.Winner()),
							Board:  build.PBBoard(g.Board),
						},
					},
				},
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
