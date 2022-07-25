package main

import (
	"math/rand"
	"regexp"
	"strconv"
	"time"

	"github.com/lonelyevil/khl"
)

var registRules []handlerRule = []handlerRule{
	{`^ping`, func(ctxCommon *khl.EventDataGeneral, s []string, f func(string) string) {
		f("来自 127.0.0.1 的回复: 字节=32 时间=" + strconv.Itoa(rand.Intn(14)) + "ms TTL=62")
	}},
	{`^\s*帮助\s*$`, func(ctxCommon *khl.EventDataGeneral, s []string, f func(string) string) {
		var str string = "当前频道支持以下命令哦\n---\n"
		str += "还没有有效命令"
		f(str)
	}},
	{`^\s*HelloWorld\s*$`, func(ctxCommon *khl.EventDataGeneral, s []string, f func(string) string) {
		if _, ok := registArray[ctxCommon.AuthorID]; ok {
			f("已在处置流程中，请尽快完成操作。")
			return
		}
		registReq(guildId, ctxCommon.AuthorID, f, localSession)
	}},
}

func registChanHandler(ctxCommon *khl.EventDataGeneral) {
	if ctxCommon.Type != khl.MessageTypeKMarkdown {
		return
	}
	reply := func(words string) string {
		resp, _ := sendMarkdown(ctxCommon.TargetID, words)
		return resp.MsgID
	}

	for n := range registRules {
		v := &registRules[n]
		r := regexp.MustCompile(v.matcher)
		matchs := r.FindStringSubmatch(ctxCommon.Content)
		if len(matchs) > 0 {
			go v.getter(ctxCommon, matchs, reply)
			return
		}
	}
}

type msdIdCode struct {
	code    string
	msgId   string
	guildId string
}

var registArray map[string]msdIdCode = make(map[string]msdIdCode)

var reactionArray = []string{
	"0️⃣",
	"1️⃣",
	"2️⃣",
	"3️⃣",
	"4️⃣",
	"5️⃣",
	"6️⃣",
	"7️⃣",
}

func registReq(guildId string, userId string, send func(string) string, session *khl.Session) {
	code := reactionArray[rand.Intn(7)]
	id := send("检测到未识别对象 (met)" + userId + "(met) ，请在`100`秒内点击本条消息下方的图标`" +
		code + "`完成消毒处置，超时或者操作错误将被强制弹出。")
	registArray[userId] = msdIdCode{code: registArray[userId].code, msgId: id, guildId: guildId}
	for _, v := range reactionArray {
		session.MessageAddReaction(id, v)
	}
	// DONE:
	// 100秒后，如果registArray[userId]仍然存在，则踢出用户，并删除条目
	go func() {
		<-time.After(time.Second * time.Duration(100))
		if _, ok := registArray[userId]; !ok {
			return
		}
		send("(met)" + userId + "(met) 超时未完成处置，已被强制弹出")
		session.GuildKickout(registArray[userId].guildId, userId)
		session.MessageDelete(registArray[userId].msgId)
		delete(registArray, userId)
	}()
}

func registJoinHandler(ctx *khl.GuildMemberAddContext) {
	send := func(words string) string {
		resp, _ := sendMarkdown(registChannel, words)
		return resp.MsgID
	}
	registReq(ctx.Common.TargetID, ctx.Extra.UserID, send, ctx.Session)
}

func registReactionHandler(ctx *khl.ReactionAddContext) {
	if ctx.Extra.UserID == botID {
		return
	}
	reply := func(words string) string {
		resp, _ := sendMarkdown(ctx.Extra.ChannelID, words)
		return resp.MsgID
	}
	go func() {
		// 判断是否是未注册用户发起的reaction
		if _, ok := registArray[ctx.Extra.UserID]; !ok {
			return
		}
		// 判断是否用户的reaction添加到了正确的msg上
		if registArray[ctx.Extra.UserID].msgId != ctx.Extra.MsgID {
			return
		}
		// 判断用户是否添加了正确的reaction
		// 如果正确，则赋予basicPrivilege权限，否则将移除用户
		if ctx.Extra.Emoji.ID == registArray[ctx.Extra.UserID].code {
			// DONE:
			ctx.Session.GuildRoleGrant(registArray[ctx.Extra.UserID].guildId, ctx.Extra.UserID, basicPrivilege)
			reply("(met)" + ctx.Extra.UserID + "(met) 已成功执行消毒处置程序，现已对其开放大厅权限")
		} else {
			// DONE:
			go func() {
				<-time.After(time.Second * time.Duration(10))
				ctx.Session.GuildKickout(registArray[ctx.Extra.UserID].guildId, ctx.Extra.UserID)
			}()
			reply("(met)" + ctx.Extra.UserID + "(met) 被识别为入侵者，将在`10`s后强制弹出")
		}
		ctx.Session.MessageDelete(registArray[ctx.Extra.UserID].msgId)
		delete(registArray, ctx.Extra.UserID)
	}()
}
