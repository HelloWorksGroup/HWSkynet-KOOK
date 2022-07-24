package main

import (
	"regexp"

	"github.com/lonelyevil/khl"
)

func sheetsSearchInWrongChannel(ctxCommon *khl.EventDataGeneral, s []string, f func(string) string) {
	f(randomDynamicSentence(sheetSearchInWrongChannel))
}

var sheetsRules []handlerRule = []handlerRule{
	{`^\s*(?:找谱|搜谱)\s+(.+)`, sheetsSearchInWrongChannel},
	{`^\s*帮助\s*$`, func(ctxCommon *khl.EventDataGeneral, s []string, f func(string) string) {
		var str string = "当前频道还没有支持的命令哦\n---\n"
		str += "求谱委托系统还在开发中哦~\n委托系统上线之后就可以方便的挂上求谱委托啦~"
		f(str)
	}},
}

func sheetChanHandler(ctxCommon *khl.EventDataGeneral) {
	if ctxCommon.Type != khl.MessageTypeText && ctxCommon.Type != khl.MessageTypeKMarkdown {
		return
	}
	if askForDmsFile(ctxCommon) {
		return
	}
	reply := func(words string) string {
		resp, _ := sendMarkdown(sheetChannel, words)
		return resp.MsgID
	}

	for n := range sheetsRules {
		v := &sheetsRules[n]
		r := regexp.MustCompile(v.matcher)
		matchs := r.FindStringSubmatch(ctxCommon.Content)
		if len(matchs) > 0 {
			go v.getter(ctxCommon, matchs, reply)
			return
		}
	}
}
