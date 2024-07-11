package main

func main() {
	addHandller("shogi", shogi)
	addHandller("azInputGameGame", azInputGameGame)
	addHandller("mashGame", mashGame)
	startServer("/test", "8444", "./fullchain.pem", "./privkey.pem")
}
