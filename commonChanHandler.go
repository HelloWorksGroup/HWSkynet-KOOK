package main

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lonelyevil/khl"
)

var commOnce sync.Once
var clockInput = make(chan interface{})
var tWords TodayWords

type handlerRule struct {
	matcher string
	getter  func(ctxCommon *khl.EventDataGeneral, matchs []string, reply func(string) string)
}

var commRules []handlerRule = []handlerRule{
	{`^YUI在么.{0,5}`, func(ctxCommon *khl.EventDataGeneral, s []string, f func(string) string) {
		if ctxCommon.AuthorID == masterID {
			f("YUI在的哦")
		}
	}},
	{`^\s*(?:找谱|搜谱)\s+(.+)`, sheetsSearchInWrongChannel},
	{`^\s*树脂\s*(\d{1,3})\s*$`, resinRemind},
	{`^\s*树脂\s*$`, resinCheck},
	{`^\s*帮助\s*$`, func(ctxCommon *khl.EventDataGeneral, s []string, f func(string) string) {
		var str string = "当前频道支持以下命令哦\n---\n"
		str += "发送 `树脂`+`数字` 如 `树脂 13` 可以跟踪树脂并自动提醒\n"
		str += "发送 `树脂` 可以查询当前的树脂量"
		f(str)
	}},
}

func resinRemind(ctxCommon *khl.EventDataGeneral, s []string, f func(string) string) {
	count, _ := strconv.Atoi(s[1])
	if count > 160 {
		f(randomSentence(resinGreaterMax))
	} else if count >= 156 {
		f(randomSentence(resinNearMax))
	} else {
		resinRecordAdd_v2(ctxCommon.AuthorID, count)
	}
}

func resinCheck(ctxCommon *khl.EventDataGeneral, s []string, f func(string) string) {
	v := resinRecordGet_v2(ctxCommon.AuthorID)
	if v != nil {
		sendMarkdown(ctxCommon.TargetID,
			"(met)"+ctxCommon.AuthorID+"(met)"+" 你现在的树脂大概有 `"+strconv.FormatInt(160-((v.EndTime-time.Now().UnixMilli())/480000), 10)+"` 点哦。")
	} else {
		sendMarkdown(ctxCommon.TargetID,
			"你还没有登记过树脂记录哦。可以用 **树脂+数字** 的语法，比如 `树脂 21` 这样来登记哦。")
	}
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

func askForDmsFile(ctxCommon *khl.EventDataGeneral) bool {
	str := strings.ToLower(ctxCommon.Content)
	if strings.Contains(str, "dms") || strings.Contains(str, "乱码") {
		sendMarkdown(ctxCommon.TargetID,
			"`dms`后缀的是加密谱面，应该使用软件上的`File`按钮导入哦。\n\n> :pig: 其实对于所有的谱面，正确的导入方式都是使用软件导入哦。\n:pig: 虽然大家老是在问这个问题，但YUI每次都会耐心回答的哦。")
		return true
	}
	return false
}

func commonChanHandler(ctxCommon *khl.EventDataGeneral) {
	if ctxCommon.Type != khl.MessageTypeText && ctxCommon.Type != khl.MessageTypeKMarkdown {
		return
	}
	if askForDmsFile(ctxCommon) {
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
