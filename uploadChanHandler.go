package main

import (
	"fmt"
	kcard "local/khlcard"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lonelyevil/khl"
)

func uploadDeleteHandler(ctx *khl.MessageDeleteContext) {
	if ctx.Extra.ChannelID != uploadChannel {
		return
	}
	for k, vc := range sheetsCache {
		if vc.Id == ctx.Extra.MsgID {
			fmt.Println("Sheets DELETE =", k)
			sheetsCache[k].Removed = true
			return
		}
	}
}

func uploadChanFileHandler(ctx *khl.FileMessageContext) {
	if ctx.Common.TargetID != uploadChannel {
		return
	}
	if ctx.Common.Type != khl.MessageTypeFile {
		return
	}
	fmt.Println("Upload ID=", ctx.Common.MsgID)
	var card kcard.KHLCard
	go func(b khl.Attachment, v *khl.EventDataGeneral) {
		afterDelete := func(msgId string, sec int) {
			<-time.After(time.Second * time.Duration(sec))
			localSession.MessageDelete(msgId)
		}
		fileExt := strings.ToLower(filepath.Ext(b.Name))
		if fileExt != ".dms" && fileExt != ".txt" {
			localSession.MessageDelete(v.MsgID)
			resp, _ := sendMarkdown(v.TargetID, "(met)"+v.AuthorID+"(met) "+randomSentence(uploadWrongType))
			go afterDelete(resp.MsgID, 15)
			return
		}
		if b.Size > 65536 {
			localSession.MessageDelete(v.MsgID)
			resp, _ := sendMarkdown(v.TargetID, "(met)"+v.AuthorID+"(met) "+randomSentence(uploadBigFile))
			go afterDelete(resp.MsgID, 15)
			return
		}
		idx := sheetAdd(sheetUpload{
			UpID:         v.AuthorID,
			Id:           v.MsgID,
			Name:         b.Name,
			Url:          b.URL,
			Type:         fileExt,
			Size:         b.Size,
			IsDownloaded: false,
		})
		if idx < 0 {
			localSession.MessageDelete(v.MsgID)
			resp, _ := sendMarkdown(v.TargetID, "(met)"+v.AuthorID+"(met) "+randomSentence(uploadDupFile))
			go afterDelete(resp.MsgID, 15)
			return
		} else {
			uploadSheet := sheetsCache[idx]
			sheetsCache[idx].FilePath = "./sheets/" + randToken(16) + uploadSheet.Type
			downloadFileTo(uploadSheet.Url, sheetsCache[idx].FilePath)
			<-time.After(time.Millisecond * time.Duration(100))
			sheetsCache[idx].IsDownloaded = true
			dupIdx, reason := isDupSheet(idx)
			if dupIdx >= 0 {
				dupSheet := sheetsCache[dupIdx]
				card = kcard.KHLCard{}
				card.Init()
				if dupSheet.UpID != uploadSheet.UpID {
					card.Card.Theme = "danger"
					card.AddModule_header(":x: 新上传的谱面未能通过 YUI 查重 :x:")
					card.AddModule_divider()
					card.AddModule_markdown("上传谱面名称:`" + uploadSheet.Name + "`\n上传者:(met)" + uploadSheet.UpID + "(met)")
					card.AddModule_markdown("与谱面:`" + dupSheet.Name + "`\n上传者:(met)" + dupSheet.UpID + "(met)\n未通过原因：" + reason)
					card.AddModule_markdown(":x: 判定上传谱面为 **重复谱面**  将予以`删除` :x:")
					sendKCard(sheetChannel, card.String())
					err := removeSheetWithIndex(idx)
					if err != nil {
						fmt.Println(err.Error())
					}
				} else {
					card.Card.Theme = "warning"
					card.AddModule_header(":up: 谱面更新 :up:")
					card.AddModule_divider()
					card.AddModule_markdown("上传谱面名称:`" + uploadSheet.Name + "`\n上传者:(met)" + uploadSheet.UpID + "(met)")
					card.AddModule_markdown("旧谱面:`" + dupSheet.Name + "`\n")
					card.AddModule_markdown(":up: 上传者更新了谱面，旧谱面已删除 :up:")
					sendKCard(sheetChannel, card.String())
					err := removeSheetWithIndex(dupIdx)
					if err != nil {
						fmt.Println(err.Error())
					}
				}
				return
			} else if dupIdx < -1 {
				fmt.Println("duplicate test error", dupIdx)
			}
		}

		var uploadCount int = 0
		for _, vc := range sheetsCache {
			if vc.UpID == v.AuthorID {
				uploadCount += 1
			}
		}
		card = kcard.KHLCard{}
		card.Init()
		card.Card.Theme = "warning"
		card.AddModule_header("YUI 又收到了新上传的谱面哦！")
		card.AddModule_divider()
		card.AddModule_markdown("**吟游诗人:** **(met)" + v.AuthorID + "(met)** *[累计上传谱面`" + strconv.Itoa(uploadCount) + "`份]*")
		card.AddModule_markdown("**伸手链接:** [" + b.Name + "](" + b.URL + ")")
		card.AddModule_markdown("**谱面类型:** `" + fileExt + "`")
		card.AddModule_markdown("**谱面大小:** " + strconv.FormatInt(b.Size, 10) + " bytes")
		sendKCard(sheetChannel, card.String())
		sheetsDbSave()
	}(ctx.Extra.Attachments, ctx.Common)
}

func uploadChanHandler(ctxCommon *khl.EventDataGeneral) {
	if ctxCommon.Type != khl.MessageTypeFile {
		if ctxCommon.AuthorID != masterID {
			resp, _ := sendMarkdown(ctxCommon.TargetID, randomSentence(uploadBan))
			botMsgID := resp.MsgID
			go func(botMsgID string) {
				<-time.After(time.Second * time.Duration(7))
				localSession.MessageDelete(botMsgID)
				localSession.MessageDelete(ctxCommon.MsgID)
			}(botMsgID)
		}
	}
}
