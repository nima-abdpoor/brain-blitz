package main

import (
	"BrainBlitz.com/game/repository/mysql"
	"BrainBlitz.com/game/service/userservice"
)

func main() {
	mysqlDB := mysql.New()
	userservice.New(mysqlDB)
}
