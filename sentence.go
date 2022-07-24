package main

//import "fmt"
import (
	"math/rand"
)

var startUp = []string{
	"YUI 上线啦！",
	"世界第一可爱的 YUI 来啦！",
	"YUI 已唤醒。",
	"YUI，启动。",
}
var newVersionStartUp = []string{
	"内核升级完毕！ " + buildVersion + "型 YUI 上线啦！",
	"升级成功！全新的 " + buildVersion + "型 YUI 现在上线！",
	"YUI变得更可爱了！" + buildVersion + "型 YUI上线！",
}

var shutdown = []string{
	"啊，YUI 要充电了呢......",
	"啊，谁把 YUI 的电源线踢掉了...",
	"YUI 的电源线好像掉了...",
	"YUI 要换个新的模块了呢",
	"YUI 现在要去换件衣服哦",
	"YUI YUI YUI YUI YUI YUI",
}
var uploadBan = []string{
	"这个频道只能上传谱面，不要发送文字信息哦。YUI来帮你删掉吧。",
	"这里是不能发送文字消息的哦，YUI马上来帮你删掉吧~",
	"这里是分享区，不能用来聊天的哦，这条消息YUI就悄悄帮你删掉了哦。",
}
var uploadWrongType = []string{
	"啊哦！这里只能上传`txt`和`dms`格式的谱面哦。YUI已经帮你删掉了错误的上传文件呢。",
}
var uploadBigFile = []string{
	"啊哦！这里只能上传不超过`64kb`的`txt`和`dms`格式的谱面哦。你上传的文件超过了大小的限制哦。如果确认你上传的文件确实是正确的谱面的话，请联系群主鉴定哦。",
}
var uploadDupFile = []string{
	"啊哦！你上传的谱面已经有人上传过了哦。YUI 已经帮你删除掉了呢。",
}
var resinGreaterMax = []string{
	"你树脂这么多，能给YUI分点么？",
}
var resinNearMax = []string{
	"你的树脂都已经快满了，就不用YUI帮你记录了。",
}
var resinFull = []string{
	"YUI戳了戳你，提醒你 (spl)树脂(spl) 已经满到快要溢出来了哦",
}
var resinAdded = []string{
	"你的期待将随时光流转而满盈。",
	"你的渴望将随日月流转而满足。",
	"你的欲望将随海潮涨落被填满。",
	"你的未来将随时针转动而到来。",
	"你的道路将随星辰大海而开辟。",
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

func sheetSearchInWrongChannel() string {
	var sentences = []string{
		"搜谱还请前往 (chn)" + searchChannel + "(chn) 频道哦，在这个频道搜谱会打扰到大家的呢。",
	}
	return sentences[rand.Intn(len(sentences))]
}

func randomDynamicSentence(fn func() string) string {
	return fn()
}

func randomSentence(list []string) string {
	return list[rand.Intn(len(list))]
}
