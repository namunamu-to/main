package main

import (
	"crypto/tls"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type player struct {
	uid, name string
	conn      *websocket.Conn
}

type room struct {
	players []player
}

func initResponseHeader(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://galleon.yachiyo.tech")
	// w.Header().Set("Access-Control-Allow-Headers", "*")
	// w.Header().Set("Access-Control-Allow-Origin", "*")
	// w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	// if r.Method == "OPTIONS" {
	// 	w.WriteHeader(http.StatusOK)
	// 	return
	// }
}

var rooms = make(map[string]*room) //roomKey

func getPlayerIdx(roomKey string, uid string) int {
	players := rooms[roomKey].players
	for i := 0; i < len(players); i++ {
		if players[i].uid == uid {
			return i
		}
	}

	return -1 //見つからなかった時
}

func enterRoom(roomKey string, player player) {
	idx := getPlayerIdx(roomKey, player.uid)
	if idx == -1 { //まだ自分が部屋に入ってなかったら追加
		rooms[roomKey].players = append(rooms[roomKey].players, player)
	}
}

func exitRoom(roomKey string, plData player) {
	pIdx := getPlayerIdx(roomKey, plData.uid)
	if pIdx != -1 { //自分が部屋にいたら部屋から抜ける
		a := rooms[roomKey].players
		a[pIdx] = a[len(a)-1]
		a = a[:len(a)-1]
		rooms[roomKey].players = a
	}

	if len(rooms[roomKey].players) == 1 { //他プレイヤーへ退出したことを通知
		sendMsg(rooms[roomKey].players[0].conn, "disConnect")
	}
}

func sendMsg(conn *websocket.Conn, msg string) {
	conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func readCmd(str string) ([]string, string, int) {
	cmd := strings.Split(string(str), " ")
	cmdType := cmd[0]
	cmdLen := len(cmd)
	return cmd, cmdType, cmdLen
}

func bloadcastMsg(roomKey string, msg string) {
	for i := 0; i < len(rooms[roomKey].players); i++ {
		sendMsg(rooms[roomKey].players[i].conn, msg)
	}
}

func isRoom(roomKey string) bool {
	_, ok := rooms[roomKey]
	return ok
}

func makeRoom(roomKey string) {
	rooms[roomKey] = &room{players: []player{}}
}

func sendMsgToAnother(roomKey string, exceptPl player, msg string) {
	//自分以外にコマンド送信
	exceptIdx := getPlayerIdx(roomKey, exceptPl.uid)
	toIdx := 1 - exceptIdx
	sendMsg(rooms[roomKey].players[toIdx].conn, msg)
}

func shogiCmd(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	//リモートアドレスからのアクセスを許可する
	conn, _ := upgrader.Upgrade(w, r, nil)

	log := time.Now().Format("2006/1/2 15:04:05") + " | " + r.RemoteAddr
	WriteFileAppend("./data/shogiAccessLog.txt", log)

	// 無限ループさせることでクライアントからのメッセージを受け付けられる状態にする
	roomKey := ""
	plData := player{uid: MakeUuid(), name: "", conn: conn}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil { //通信終了時の処理
			if roomKey == "" { //
				break
			}

			exitRoom(roomKey, plData) //部屋から抜ける
			break
		}

		//msgのコマンド読み取り
		cmd, cmdType, cmdLen := readCmd(string(msg))

		//コマンドに応じた処理をする
		if cmdType == "roomMatch" && cmdLen == 4 { //マッチングコマンド。想定コマンド = "roomMatch 部屋番号 部屋パスワード プレイヤー名"
			roomKey = cmd[1] + cmd[2]

			if !isRoom(roomKey) { //部屋が無いなら作る
				makeRoom(roomKey)
			}

			playerNum := len(rooms[roomKey].players)

			if playerNum == 2 { //部屋がいっぱいなら
				sendMsg(conn, "fullMember")

			} else if playerNum < 2 { //人数が揃ってないとき
				//部屋に入る
				enterRoom(roomKey, plData)
				playerNum = len(rooms[roomKey].players)

				if playerNum == 2 { //人数が揃った時
					bloadcastMsg(roomKey, "matched")
				} else if playerNum == 1 {
					bloadcastMsg(roomKey, "matching "+strconv.Itoa(playerNum))
				}

			}

		} else if cmdType == "move" && cmdLen == 5 { //移動コマンド。想定コマンド = "move pieceId toX toY reverse"
			sendMsgToAnother(roomKey, plData, string(msg))
		}
	}
}

func main() {
	// ハンドラの設定
	mux := http.NewServeMux() //ミューテックス。すでに起動してるか確認。
	mux.HandleFunc("/shogi/cmd", shogiCmd)
	mux.HandleFunc("/mashGame/cmd", MashGameCmd)
	mux.HandleFunc("/azInputGame/cmd", azInputGameGameCmd)

	//tls設定
	cfg := &tls.Config{
		ClientAuth: tls.RequestClientCert,
	}

	//サーバー設定
	srv := http.Server{
		Addr:      ":8443",
		Handler:   mux,
		TLSConfig: cfg,
	}

	// http.HandleFunc("/postRequest", handler)
	println("サーバー起動")
	// err := srv.ListenAndServeTLS("/etc/letsencrypt/live/os3-382-24260.vs.sakura.ne.jp/fullchain.pem", "/etc/letsencrypt/live/os3-382-24260.vs.sakura.ne.jp/privkey.pem")
	err := srv.ListenAndServeTLS("./fullchain.pem", "./privkey.pem")
	if err != nil {
		println(err.Error())
	}
}
