package main

import (
	"net/http"
	"strconv"
	"time"
)

var azInputGameFiles = map[string]string{
	"ranking":   "./data/azInputGameRanking.csv",
	"accessLog": "./data/azInputGameAccessLog.txt",
}

var azInputGameRankingData = ReadCsv(azInputGameFiles["ranking"])
var azList = "abcdefghijklmnopqrstuvwxyz" //a～zのリスト

// ランキング更新
func updateAzGameRanking(userName string, newScore int) int {
	addData := []string{userName, strconv.Itoa(newScore)}
	ranking := len(azInputGameRankingData)
	for i, line := range azInputGameRankingData {
		lineScore, _ := strconv.Atoi(line[1])

		if lineScore < newScore {
			continue
		}

		ranking = i
		break
	}

	slice1 := azInputGameRankingData[:ranking]
	slice2 := [][]string{addData}
	slice3 := azInputGameRankingData[ranking:]
	slice2 = append(slice2, slice3...)
	azInputGameRankingData = append(slice1, slice2...)
	WriteCsv(azInputGameFiles["ranking"], azInputGameRankingData)
	return ranking + 1
}

func azInputGameGameCmd(w http.ResponseWriter, r *http.Request) {
	initResponseHeader(w, r)
	// upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	//リモートアドレスからのアクセスを許可する
	conn, _ := upgrader.Upgrade(w, r, nil)

	log := time.Now().Format("2006/1/2 15:04:05") + " | " + r.RemoteAddr
	WriteFileAppend(azInputGameFiles["accessLog"], log)

	//プレイ中の情報
	playing := false
	nextKey := string(azList[0])
	nextIdx := 0
	elapsedTime := 0
	name := "名無し"
	myRank := len(azInputGameRankingData)

	// 無限ループさせ、接続が切れないようにする
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil { //通信終了時の処理
			break
		}

		//msgのコマンド読み取り
		cmd, _, cmdLen := readCmd(string(msg))

		if cmd[0] == "startGame" && cmdLen == 2 { //ゲーム開始コマンド。想定コマンド = startGame userName
			if cmd[1] != "" { //名前が空じゃなかったら、名前を更新
				name = cmd[1]
			}

			if !playing {
				go func() { //ゲーム中の処理
					playing = true
					nextKey = string(azList[0])
					elapsedTime = 0
					nextIdx = 0

					timer := time.NewTicker(time.Duration(1) * time.Millisecond)
					for {
						<-timer.C
						elapsedTime++
						if nextIdx == len(azList) { //プレイが終わったら次のプレイ準備をし、スコアの処理を行う
							playing = false
							myRank = updateAzGameRanking(name, elapsedTime)
							sendMsg(conn, "rankingData "+strconv.Itoa(myRank)+" "+strconv.Itoa(elapsedTime)+" "+SliceToCsvStr(azInputGameRankingData[:5]))
							break
						}
					}
				}()
			}

		} else if cmd[0] == "keyDown" && cmdLen == 2 { //連打ボタンコマンド。想定コマンド = keyDown key
			if playing {
				if cmd[1] == nextKey {
					nextIdx++
					if nextIdx != len(azList) {
						nextKey = string(azList[nextIdx])
					}
				} else {
					elapsedTime += 1000
				}
			}
		} else if cmd[0] == "getRanking" && cmdLen == 1 { //ランキング取得コマンド。想定コマンド = getRanking
			sendMsg(conn, "rankingData "+strconv.Itoa(myRank)+" "+strconv.Itoa(elapsedTime)+" "+SliceToCsvStr(azInputGameRankingData[:5]))
		}
	}
}
