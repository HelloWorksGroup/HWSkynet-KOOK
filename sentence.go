package main

//import "fmt"
import (
	"math/rand"
)

var shutdown = []string{
	"Skynet Rebooting......",
	"Skynet Skynet Skynet Skynet Skynet Skynet",
}
var idle = []string{
	"...",
	"。。。",
	"？？？",
	"2333",
	"现充！",
	"不开心",
	"。。。。",
	"。。",
}

func randomDynamicSentence(fn func() string) string {
	return fn()
}

func randomSentence(list []string) string {
	return list[rand.Intn(len(list))]
}
