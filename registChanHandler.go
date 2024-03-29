package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"github.com/lonelyevil/kook"
)

var registRules []handlerRule = []handlerRule{
	{`^ping`, func(ctxCommon *kook.EventDataGeneral, s []string, f func(string) string) {
		f(randomDynamicSentence(pong))
	}},
	{`^\s*帮助\s*$`, func(ctxCommon *kook.EventDataGeneral, s []string, f func(string) string) {
		var str string = "当前频道支持以下命令哦\n---\n"
		str += "还没有有效命令"
		f(str)
	}},
	{`^\s*HelloWorld\s*$`, func(ctxCommon *kook.EventDataGeneral, s []string, f func(string) string) {
		if _, ok := registArray[ctxCommon.AuthorID]; ok {
			f("已在处置流程中，请尽快完成操作。")
			return
		}
		registReq(gGuildId, ctxCommon.AuthorID, f, localSession)
	}},
}

func registChanHandler(ctxCommon *kook.EventDataGeneral) {
	if ctxCommon.Type != kook.MessageTypeKMarkdown {
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

func registReq(guildId string, userId string, send func(string) string, session *kook.Session) {
	code := reactionArray[rand.Intn(7)]
	id := send("检测到未识别对象 (met)" + userId + "(met) ，请在`100`秒内点击本条消息下方的图标`" +
		code + "`完成消毒处置，超时或者操作错误将 ~~被强制弹出~~ 需要重新进行消毒。")
	registArray[userId] = msdIdCode{code: code, msgId: id, guildId: guildId}
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
		send("(met)" + userId + "(met) 超时未完成处置，可以发送`HelloWorld`重新进行消毒程序。")
		fmt.Println("Kick", userId, "from", guildId, "origin", registArray[userId].guildId)
		// session.GuildKickout(guildId, userId)
		session.MessageDelete(registArray[userId].msgId)
		delete(registArray, userId)
	}()
}

func registJoinHandler(ctx *kook.GuildMemberAddContext) {
	send := func(words string) string {
		resp, _ := sendMarkdown(registChannel, words)
		return resp.MsgID
	}
	u, _ := ctx.Session.UserView(ctx.Extra.UserID, kook.UserViewWithGuildID(ctx.Common.TargetID))
	if u.MobileVerified {
		registReq(ctx.Common.TargetID, ctx.Extra.UserID, send, ctx.Session)
	} else {
		sendMarkdown(registChannel, "(met)"+ctx.Extra.UserID+"(met) 由于未通过手机认证，无法启动消毒程序。"+
			"\n通过手机认证后，可以发送`HelloWorld`手动触发消毒程序。")
	}
}

func registReactionHandler(ctx *kook.ReactionAddContext) {
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
			_, err := ctx.Session.GuildRoleGrant(registArray[ctx.Extra.UserID].guildId, ctx.Extra.UserID, basicPrivilege)
			if err != nil {
				reply("(met)" + ctx.Extra.UserID + "(met) 授予权限失败，可能原因是用户未通过实名认证")
			} else {
				reply("(met)" + ctx.Extra.UserID + "(met) 已成功执行消毒处置程序，已为你开放大厅权限。" +
					"\n现在可以进入 (chn)4641712772375343(chn) 小憩。")
			}
		} else {
			// DONE:
			// go func() {
			// 	<-time.After(time.Second * time.Duration(10))
			// 	fmt.Println("Kick", ctx.Extra.UserID, "from", guildId, "origin", registArray[ctx.Extra.UserID].guildId)
			// 	ctx.Session.GuildKickout(guildId, ctx.Extra.UserID)
			// }()
			fmt.Println("want", registArray[ctx.Extra.UserID].code, "get", ctx.Extra.Emoji.ID)
			reply("(met)" + ctx.Extra.UserID + "(met) 消毒处置失败，可以发送`HelloWorld`重新进行消毒程序。")
		}
		ctx.Session.MessageDelete(registArray[ctx.Extra.UserID].msgId)
		delete(registArray, ctx.Extra.UserID)
	}()
}
