package main

import (
	"regexp"
	"sort"
	"strconv"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/lonelyevil/khl"
)

var searchRules []handlerRule = []handlerRule{
	{`^\s*(?:找谱|搜谱)\s+(.+)`, sheetsSearch},
	{`^\s*帮助\s*$`, func(ctxCommon *khl.EventDataGeneral, s []string, f func(string) string) {
		var str string = "当前频道支持以下命令哦\n---\n"
		str += "发送 `找谱`或者`搜谱` + `谱面关键字` 如 `搜谱 起风` 可以搜索库中的谱面\n"
		f(str)
	}},
}

func sheetsSearch(ctxCommon *khl.EventDataGeneral, s []string, f func(string) string) {
	var searchSlice []string
	if len([]rune(s[1])) < 2 {
		go f("搜索关键词的长度至少要达到`2`哦")
		return
	}
	for _, v := range sheetsCache {
		if v.Removed {
			searchSlice = append(searchSlice, "")
		} else {
			searchSlice = append(searchSlice, v.Name)
		}
	}
	matches := fuzzy.RankFindNormalizedFold(s[1], searchSlice)
	sort.Sort(matches)
	var replyWords string = ""
	if len(matches) == 0 {
		replyWords += ":x: 非常抱歉，YUI **没有**能找到你要的谱呢 :x:"
	} else if len(matches) <= 7 {
		replyWords += ":white_check_mark: YUI 帮你找到了下面这些谱哦\n---\n"
		for _, v := range matches {
			replyWords += ":musical_score: [" + sheetsCache[v.OriginalIndex].Name + "](" + sheetsCache[v.OriginalIndex].Url + ")"
			replyWords += " - 上传者:(met)" + sheetsCache[v.OriginalIndex].UpID + "(met)\n"
		}
	} else {
		replyWords += ":white_check_mark: YUI 一共找到了`" + strconv.Itoa(len(matches)) + "`份相关的谱面，列出名称最相似的5个如下：\n---\n"
		for i := 0; i < 5; i++ {
			replyWords += ":musical_score: [" + sheetsCache[matches[i].OriginalIndex].Name + "](" + sheetsCache[matches[i].OriginalIndex].Url + ")"
			replyWords += " - 上传者:(met)" + sheetsCache[matches[i].OriginalIndex].UpID + "(met)\n"
		}
	}
	go f(replyWords)
}

func searchChanHandler(ctxCommon *khl.EventDataGeneral) {
	if ctxCommon.Type != khl.MessageTypeText && ctxCommon.Type != khl.MessageTypeKMarkdown {
		return
	}
	reply := func(words string) string {
		resp, _ := sendMarkdown(searchChannel, words)
		return resp.MsgID
	}

	for n := range searchRules {
		v := &searchRules[n]
		r := regexp.MustCompile(v.matcher)
		matchs := r.FindStringSubmatch(ctxCommon.Content)
		if len(matchs) > 0 {
			go v.getter(ctxCommon, matchs, reply)
			return
		}
	}
}
