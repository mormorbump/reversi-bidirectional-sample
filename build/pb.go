package build

import (
	"kazuki.matsumoto/reversi/game"
	"kazuki.matsumoto/reversi/gen/pb"
)

func PBRoom(r *game.Room) *pb.Room {
	return &pb.Room{
		Id:    r.ID,
		Host:  PBPlayer(r.Host),
		Guest: PBPlayer(r.Guest),
	}
}

func PBPlayer(p *game.Player) *pb.Player {
	if p == nil {
		return nil
	}
	return &pb.Player{
		Id:        p.ID,
		Character: PBCharacter(p.Character),
	}
}

func PBCharacter(c game.Character) pb.Character {
	switch c {
	case game.Black:
		return pb.Character_BLACK
	case game.White:
		return pb.Character_WHITE
	case game.Empty:
		return pb.Character_EMPTY
	case game.Wall:
		return pb.Character_WALL
	}
	return pb.Character_UNKNOWN
}

func PBBoard(b *game.Board) *pb.Board {
	// 列
	pbCols := make([]*pb.Board_Col, 0, 10)
	// protobufで二次元配列を直接扱えないので、cellの数 -> colの数ぶんpbCellsを定義。
	for _, col := range b.Cells {
		pbCells := make([]pb.Character, 0, 10)
		// colも同様に行列を持つので、その数分配列を生成。要素をpbCellsにappendしていく。
		for _, c := range col {
			pbCells = append(pbCells, PBCharacter(c))
		}
		// pbColsとpbCellsを結合し、二次元配列とする
		pbCols = append(pbCols, &pb.Board_Col{
			Cells: pbCells,
		})
	}
	return &pb.Board{Cols: pbCols}
}
