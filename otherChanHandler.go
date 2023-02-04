package main

import (
	"regexp"

	"github.com/lonelyevil/kook"
)

func otherChanHandler(ctxCommon *kook.EventDataGeneral) {
	if ctxCommon.Type != kook.MessageTypeKMarkdown {
		return
	}
	reply := func(words string) string {
		resp, _ := sendMarkdown(ctxCommon.TargetID, words)
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
