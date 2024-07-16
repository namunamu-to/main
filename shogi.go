package main

import (
	"strconv"
)

func shogi(plData player) {
	for {
		_, msg, err := plData.conn.ReadMessage()
		if err != nil { //通信終了時の処理
			exitRoom(plData.roomKey, &plData) //部屋から抜ける

			break
		}

		//msgのコマンド読み取り
		_, cmdType, cmdLen := readCmd(string(msg))

		//コマンドに応じた処理をする
		if cmdType == "moveRoom" && cmdLen == 3 { //マッチングコマンド。想定コマンド = "moveRoom roomKey プレイヤー名"
			if !isRoom(plData.roomKey) { //部屋が無いなら作る
				makeRoom(plData.roomKey)
			}

			playerNum := len(rooms[plData.roomKey].players)

			if playerNum == 2 { //部屋がいっぱいなら
				sendMsg(plData.conn, "fullMember")

			} else if playerNum < 2 { //人数が揃ってないとき
				//部屋に入る
				enterRoom(plData.roomKey, &plData)
				playerNum = len(rooms[plData.roomKey].players)

				if playerNum == 2 { //人数が揃った時
					bloadcastMsg(plData.roomKey, "matched")
				} else if playerNum == 1 {
					bloadcastMsg(plData.roomKey, "matching "+strconv.Itoa(playerNum))
				}

			}

		} else if cmdType == "move" && cmdLen == 5 { //移動コマンド。想定コマンド = "move pieceId toX toY reverse"
			sendMsgToAnother(plData.roomKey, plData, string(msg))
		}
	}
}
