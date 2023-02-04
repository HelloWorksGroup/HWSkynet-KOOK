package main

import (
	"regexp"
	"sync"
	"time"

	"github.com/lonelyevil/kook"
)

var commOnce sync.Once
var clockInput = make(chan interface{})
var tWords TodayWords

var commRules []handlerRule = []handlerRule{
	{`^ping`, func(ctxCommon *kook.EventDataGeneral, s []string, f func(string) string) {
		f(randomDynamicSentence(pong))
	}},
	{`^\s*帮助\s*$`, func(ctxCommon *kook.EventDataGeneral, s []string, f func(string) string) {
		var str string = "当前频道支持以下命令哦\n---\n"
		str += "还没有有效命令"
		f(str)
	}},
}

func commonChanHandlerInit() {
	commOnce.Do(func() { go clock(clockInput) })
	tWords.NewDay()
}

func clock(input chan interface{}) {
	min := time.NewTicker(1 * time.Minute)
	halfhour := time.NewTicker(23 * time.Minute)
	for {
		select {
		case <-min.C:
			hour, min, _ := time.Now().Local().Clock()
			if min == 0 && hour == 5 {
				tWords.NewDay()
			}
			if !isTodayWakeuped() && !tWords.IsTodaySaid && hour >= 9 {
				words := tWords.TrySay()
				if len(words.Quote) > 0 {
					setWakeup()
					sendKCard(commonChannel, tWords.MakeKHLCard())
				}
			}
		case <-halfhour.C:
		}
	}
}

func commonChanHandler(ctxCommon *kook.EventDataGeneral) {
	if ctxCommon.Type != kook.MessageTypeKMarkdown {
		return
	}
	reply := func(words string) string {
		resp, _ := sendMarkdown(commonChannel, words)
		return resp.MsgID
	}

	for n := range commRules {
		v := &commRules[n]
		r := regexp.MustCompile(v.matcher)
		matchs := r.FindStringSubmatch(ctxCommon.Content)
		if len(matchs) > 0 {
			go v.getter(ctxCommon, matchs, reply)
			return
		}
	}
}
