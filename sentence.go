package main

//import "fmt"
import (
	"math/rand"
	"strconv"
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

func pong() string {
	return "来自 `127.0.0.1` 的回复: 字节=`32` 时间=`" + strconv.Itoa(rand.Intn(14)) + "ms` TTL=`62`"
}

func randomDynamicSentence(fn func() string) string {
	return fn()
}

func randomSentence(list []string) string {
	return list[rand.Intn(len(list))]
}
