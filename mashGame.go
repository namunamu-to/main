package main

import (
	"strconv"
	"time"
)

func mashGame(plData player) {
	// プレイ中の情報
	playing := false
	nowPushCount := 0
	timeLimit := 10
	name := "名無し"
	myRank := len(rankingData)

	for {
		_, msgByte, err := plData.conn.ReadMessage()
		if err != nil { //通信終了時の処理
			break
		}

		cmd, cmdType, cmdLen := readCmd(string(msgByte))

		if cmdType == "startGame" && cmdLen == 2 { //ゲーム開始コマンド。想定コマンド = startGame userName
			if cmd[1] != "" { //名前が空じゃなかったら、名前を更新
				name = cmd[1]
			}

			if !playing {
				timeLimit = 10
				playing = true
				nowPushCount = 0

				//ゲーム中の処理
				go func() {
					timeLimit = 10
					timer := time.NewTicker(time.Duration(1) * time.Second)
					for {
						<-timer.C
						timeLimit--
						if timeLimit == 0 { //プレイが終わったら次のプレイ準備をし、スコアの処理を行う
							playing = false
							myRank = updateRanking(name, nowPushCount)
							sendMsg(plData.conn, "rankingData "+strconv.Itoa(myRank)+" "+SliceToCsvStr(rankingData[:5]))
							return
						}
					}
				}()
			}

		} else if cmdType == "pushBtn" && cmdLen == 1 { //連打ボタンコマンド。想定コマンド = pushBtn
			if playing {
				nowPushCount += 1
			}
		} else if cmdType == "getRanking" && cmdLen == 1 { //ランキング取得コマンド。想定コマンド = getRanking 自分のスコア
			sendMsg(plData.conn, "rankingData "+strconv.Itoa(myRank)+" "+SliceToCsvStr(rankingData[:5]))
		}

	}
}

type Files struct {
	ranking   string
	accessLog string
}

var files = Files{
	ranking:   "./data/mashGameRanking.csv",
	accessLog: "./data/mashGameAccessLog.txt",
}

var rankingData = ReadCsv(files.ranking)

// ランキング更新
func updateRanking(userName string, newScore int) int {
	addData := []string{userName, strconv.Itoa(newScore)}
	ranking := len(rankingData)
	for i, line := range rankingData {
		lineScore, _ := strconv.Atoi(line[1])

		if lineScore > newScore {
			continue
		}

		ranking = i
		break
	}

	slice1 := rankingData[:ranking]
	slice2 := [][]string{addData}
	slice3 := rankingData[ranking:]
	slice2 = append(slice2, slice3...)
	rankingData = append(slice1, slice2...)
	WriteCsv(files.ranking, rankingData)
	return ranking + 1
}
