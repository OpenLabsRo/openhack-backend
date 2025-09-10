package utils

import (
	"math/rand"
	"strconv"
)

var Encoding string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
var TeamIDEncoding string = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
var CodeEncoding string = "abcdefghijklmnopqrstuvwxyz0123456789"

func GenTeamID() (ID string) {
	for range 6 {
		ID += string(TeamIDEncoding[rand.Intn(26)])
	}

	return
}

func GenID(n int) string {
	var ID string

	for range n {
		ID += string(Encoding[rand.Intn(64)])
	}

	return ID
}

func GenCode(n int) string {
	var code string

	for range n {
		code += strconv.Itoa(rand.Intn(10))
	}

	return code
}
