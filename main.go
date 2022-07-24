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
	scribble "github.com/nanobox-io/golang-scribble"
	"github.com/phuslu/log"
	"github.com/spf13/viper"
)

func buildUpdateLog() string {
	return "添加了在错误的频道搜谱的回复。大家搜谱请前往搜谱专用频道 (chn)" + searchChannel + "(chn) 哦！不想被打扰可以直接静音搜谱频道哦！\n更多谱面管理功能还在开发中。对于YUI的功能欢迎大家提供更多建议哦。"
}

var buildVersion string = "YUI-One Alpha0029"
var commonChannel string
var searchChannel string
var uploadChannel string
var sheetChannel string
var isVersionChange bool = false
var lastWakeupDay string // 上一次唤醒日期，用于限定每日一次的输出
var masterID string
var botID string

var localSession *khl.Session

var db *scribble.Driver

func isTodayWakeuped() bool {
	if lastWakeupDay == strconv.Itoa(time.Now().Local().Day()) {
		return true
	}
	return false
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
	db, _ = scribble.New("./database", nil)

	viper.SetDefault("token", "0")
	viper.SetDefault("commonChannel", "0")
	viper.SetDefault("uploadChannel", "0")
	viper.SetDefault("sheetChannel", "0")
	viper.SetDefault("searchChannel", "0")
	viper.SetDefault("lastWakeupDay", "0")
	viper.SetDefault("masterID", "")
	viper.SetDefault("lastwordsID", "")
	viper.SetDefault("oldversion", "0.0.0")
	viper.SetConfigType("json")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	masterID = viper.Get("masterID").(string)
	commonChannel = viper.Get("commonChannel").(string)
	uploadChannel = viper.Get("uploadChannel").(string)
	sheetChannel = viper.Get("sheetChannel").(string)
	searchChannel = viper.Get("searchChannel").(string)
	lastWakeupDay = viper.Get("lastWakeupDay").(string)
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
	s.AddHandler(txtMessageHandler)
	s.AddHandler(markdownMessageHandler)
	s.AddHandler(uploadChanFileHandler)
	s.AddHandler(uploadDeleteHandler)
	s.AddHandler(statuHan)
	s.Open()
	localSession = s
	commonChanHandlerInit()
	sheetManagerInit()
	resinRecordInit()

	if isVersionChange {
		go func() {
			<-time.After(time.Second * time.Duration(3))
			card := kcard.KHLCard{}
			card.Init()
			card.Card.Theme = "success"
			card.AddModule_header("YUI 热更新完成")
			card.AddModule_divider()
			card.AddModule_markdown("当前版本号：`" + buildVersion + "`")
			card.AddModule_markdown("**更新内容：**\n" + buildUpdateLog())
			sendKCard(commonChannel, card.String())
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

	lastResp, _ := sendMarkdown(commonChannel, randomSentence(shutdown))

	viper.Set("lastwordsID", lastResp.MsgID)
	fmt.Println("[Write] lastwordsID=", lastResp.MsgID)
	viper.WriteConfig()
	sheetsDbSave()
	fmt.Println("Bot will shutdown after 1 second.")

	<-time.After(time.Second * time.Duration(1))
	// Cleanly close down the KHL session.
	s.Close()
}

func main() {
	overseer.Run(overseer.Config{
		Required: true,
		Program:  prog,
		Fetcher:  &fetcher.File{Path: "YUI-KHL"},
		Debug:    false,
	})
}

// FileMessageContext
// MessageButtonClickContext
func statuHan(ctx *khl.BotJoinContext) {
	fmt.Println("Bot Join:" + ctx.Extra.GuildID)
}
func markdownMessageHandler(ctx *khl.KmarkdownMessageContext) {
	if ctx.Extra.Author.Bot {
		return
	}
	switch ctx.Common.TargetID {
	case botID:
		directMessageHandler(ctx.Common)
	case commonChannel:
		commonChanHandler(ctx.Common)
	case uploadChannel:
		uploadChanHandler(ctx.Common)
	case sheetChannel:
		sheetChanHandler(ctx.Common)
	case searchChannel:
		searchChanHandler(ctx.Common)
	}
}
func txtMessageHandler(ctx *khl.TextMessageContext) {
	if ctx.Extra.Author.Bot {
		return
	}
	switch ctx.Common.TargetID {
	case botID:
		directMessageHandler(ctx.Common)
	case commonChannel:
		commonChanHandler(ctx.Common)
	case uploadChannel:
		uploadChanHandler(ctx.Common)
	case sheetChannel:
		sheetChanHandler(ctx.Common)
	case searchChannel:
		searchChanHandler(ctx.Common)
	}
}
