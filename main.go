package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	"math/rand"

	kcard "local/khlcard"

	"github.com/jpillora/overseer"
	"github.com/jpillora/overseer/fetcher"
	"github.com/lonelyevil/khl"
	"github.com/lonelyevil/khl/log_adapter/plog"
	"github.com/phuslu/log"
	"github.com/spf13/viper"
)

func buildUpdateLog() string {
	return "测试更新，添加认证提示。添加频道跳转提示。\n\nHelloWorks-Skynet@[GitHub](https://github.com/Nigh/HWSkynet-KOOK)"
}

var buildVersion string = "Skynet Alpha0007"

// TODO:
// 未找到合适方法在消息事件的上下文中获取服务器ID，暂时写这里了
var guildId string = "6067588674873845"

// stdout
var stdoutChannel string

// 消毒室
var registChannel string

// 茶室频道
var commonChannel string

// 电玩房频道
var gameChannel string

// 特工频道
var ingressChannel string

// 基础权限ID
var basicPrivilege int64

type handlerRule struct {
	matcher string
	getter  func(ctxCommon *khl.EventDataGeneral, matchs []string, reply func(string) string)
}

var isVersionChange bool = false
var lastWakeupDay string // 上一次唤醒日期，用于限定每日一次的输出
var masterID string
var botID string

var localSession *khl.Session

func isTodayWakeuped() bool {
	return lastWakeupDay == strconv.Itoa(time.Now().Local().Day())
}

func setWakeup() {
	lastWakeupDay = strconv.Itoa(time.Now().Local().Day())
	viper.Set("lastWakeupDay", lastWakeupDay)
	viper.WriteConfig()
}

func sendKCard(target string, content string) (resp *khl.MessageResp, err error) {
	return localSession.MessageCreate((&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     khl.MessageTypeCard,
			TargetID: target,
			Content:  content,
		},
	}))
}
func sendMarkdown(target string, content string) (resp *khl.MessageResp, err error) {
	return localSession.MessageCreate((&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     khl.MessageTypeKMarkdown,
			TargetID: target,
			Content:  content,
		},
	}))
}

func sendMarkdownDirect(target string, content string) (mr *khl.MessageResp, err error) {
	return localSession.DirectMessageCreate(&khl.DirectMessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     khl.MessageTypeKMarkdown,
			TargetID: target,
			Content:  content,
		},
	})
}

func prog(state overseer.State) {
	fmt.Printf("App#[%s] start ...\n", state.ID)
	rand.Seed(time.Now().UnixNano())

	viper.SetDefault("token", "0")

	viper.SetDefault("stdoutChannel", "0")
	viper.SetDefault("registChannel", "0")
	viper.SetDefault("commonChannel", "0")
	viper.SetDefault("gameChannel", "0")
	viper.SetDefault("ingressChannel", "0")
	viper.SetDefault("basicPrivilege", "4848723")
	viper.SetDefault("lastWakeupDay", "0")
	viper.SetDefault("masterID", "")
	viper.SetDefault("lastwordsID", "")
	viper.SetDefault("oldversion", "0.0.0")
	viper.SetConfigType("json")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
	masterID = viper.Get("masterID").(string)
	stdoutChannel = viper.Get("stdoutChannel").(string)
	registChannel = viper.Get("registChannel").(string)
	commonChannel = viper.Get("commonChannel").(string)
	gameChannel = viper.Get("gameChannel").(string)
	ingressChannel = viper.Get("ingressChannel").(string)
	lastWakeupDay = viper.Get("lastWakeupDay").(string)
	basicPrivilege, _ = strconv.ParseInt(viper.Get("basicPrivilege").(string), 10, 64)
	if viper.Get("oldversion").(string) != buildVersion {
		isVersionChange = true
	}

	viper.Set("oldversion", buildVersion)

	l := log.Logger{
		Level:  log.InfoLevel,
		Writer: &log.ConsoleWriter{},
	}
	token := viper.Get("token").(string)
	fmt.Println("token=" + token)

	s := khl.New(token, plog.NewLogger(&l))
	me, _ := s.UserMe()
	fmt.Println("ID=" + me.ID)
	botID = me.ID
	s.AddHandler(markdownMessageHandler)
	s.AddHandler(registJoinHandler)
	s.AddHandler(registReactionHandler)

	s.Open()
	localSession = s
	commonChanHandlerInit()

	if isVersionChange {
		go func() {
			<-time.After(time.Second * time.Duration(3))
			card := kcard.KHLCard{}
			card.Init()
			card.Card.Theme = "success"
			card.AddModule_header("Skynet 热更新完成")
			card.AddModule_divider()
			card.AddModule_markdown("当前版本号：`" + buildVersion + "`")
			card.AddModule_markdown("**更新内容：**\n" + buildUpdateLog())
			sendKCard(stdoutChannel, card.String())
		}()
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.")

	fmt.Println("[Read] lastwordsID=", viper.Get("lastwordsID").(string))
	if viper.Get("lastwordsID").(string) != "" {
		go func() {
			<-time.After(time.Second * time.Duration(7))
			s.MessageDelete(viper.Get("lastwordsID").(string))
			viper.Set("lastwordsID", "")
			viper.WriteConfig()
		}()
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, overseer.SIGUSR2)
	<-sc

	lastResp, _ := sendMarkdown(stdoutChannel, randomSentence(shutdown))

	viper.Set("lastwordsID", lastResp.MsgID)
	fmt.Println("[Write] lastwordsID=", lastResp.MsgID)
	viper.WriteConfig()
	fmt.Println("Bot will shutdown after 1 second.")

	<-time.After(time.Second * time.Duration(1))
	// Cleanly close down the KHL session.
	s.Close()
}

func main() {
	overseer.Run(overseer.Config{
		Required: true,
		Program:  prog,
		Fetcher:  &fetcher.File{Path: "HWSkynet-KOOK"},
		Debug:    false,
	})
}

// FileMessageContext
// MessageButtonClickContext
func markdownMessageHandler(ctx *khl.KmarkdownMessageContext) {
	if ctx.Extra.Author.Bot {
		return
	}
	switch ctx.Common.TargetID {
	case botID:
		directMessageHandler(ctx.Common)
	case registChannel:
		registChanHandler(ctx.Common)
	case commonChannel:
		commonChanHandler(ctx.Common)
	case gameChannel:
		otherChanHandler(ctx.Common)
	case ingressChannel:
		otherChanHandler(ctx.Common)
	}
}
