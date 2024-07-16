package main

func main() {
	addHandller("shogi", "/commonGameServer/shogi", shogi)
	addHandller("azInputGame", "/commonGameServer/azInputGame", azInputGameGame)
	addHandller("mashGame", "/commonGameServer/mashGame", mashGame)
	startServer("8444", "./fullchain.pem", "./privkey.pem")
}
