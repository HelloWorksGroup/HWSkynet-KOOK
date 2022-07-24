package main

import (
	kcard "local/khlcard"
	"strconv"
	"sync"
	"time"

	"github.com/lonelyevil/khl"
	scribble "github.com/nanobox-io/golang-scribble"
)

var resinOnce sync.Once

type resinNotice struct {
	Id   string
	Type string
}

type resinRecordv2 struct {
	Id      string
	EndTime int64
	Notices []resinNotice
}

var resinRecordv2Cache []resinRecordv2

var resinClockInput = make(chan interface{})

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func resinClock(input chan interface{}) {
	minute := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-minute.C:
			resinRecordUpdate_v2()
		}
	}
}

func resinRecordInit() {
	db, _ = scribble.New("./database", nil)
	db.Read("db", "resinv2", &resinRecordv2Cache)
	resinOnce.Do(func() { go resinClock(resinClockInput) })
}

func resinFullTimeFromCount(count int) int64 {
	return time.Now().UnixMilli() + int64((160-count)*480000)
}
func resinCountFromTime(endTime int64) int {
	return min(160, int(160-(endTime-time.Now().UnixMilli())/480000))
}
func resinRecordGet_v2(id string) *resinRecordv2 {
	for _, v := range resinRecordv2Cache {
		if v.Id == id {
			return &v
		}
	}
	return nil
}

func resinRecordAdd_v2(id string, resinCount int) {

	endTime := resinFullTimeFromCount(resinCount)

	var firstLine string = "(met)" + id + "(met) 树脂"
	alreadyRecord := false
	var recordIdx int
	for i, v := range resinRecordv2Cache {
		if v.Id == id {
			firstLine += "由`" + strconv.Itoa(resinCountFromTime(v.EndTime)) + "`点更新"
			for _, b := range v.Notices {
				if b.Type == "countdown" || b.Type == "first" {
					localSession.MessageDelete(b.Id)
				}
			}
			resinRecordv2Cache[i].Notices = []resinNotice{}
			resinRecordv2Cache[i].EndTime = endTime
			alreadyRecord = true
			recordIdx = i
			break
		}
	}
	if !alreadyRecord {
		firstLine += "登记"
		resinRecordv2Cache = append(resinRecordv2Cache, resinRecordv2{id, endTime, []resinNotice{}})
		recordIdx = len(resinRecordv2Cache) - 1
	}
	firstLine += "为`" + strconv.Itoa(resinCount) + "`点"

	card := kcard.KHLCard{}
	card.Init()
	card.Card.Theme = "info"
	card.AddModule_markdown(firstLine)
	card.AddModule_markdown(randomSentence(resinAdded))
	card.AddModule_markdown("你的树脂将在如下时间后满盈：")
	card.AddModule_divider()
	card.AddModule(
		kcard.KModule{
			Type:    "countdown",
			Mode:    "day",
			EndTime: endTime,
		},
	)
	resp, _ := sendKCard(commonChannel, card.String())
	resinRecordv2Cache[recordIdx].Notices = append(resinRecordv2Cache[recordIdx].Notices, resinNotice{resp.MsgID, "first"})
	return
}

func resinRecordSaveLocal_v2() error {
	return db.Write("db", "resinv2", resinRecordv2Cache)
}

func resinRecordUpdate_v2() {
	var validIdx int = 0
	for _, v := range resinRecordv2Cache {
		timeLeft := v.EndTime - time.Now().UnixMilli()
		if timeLeft <= 0 {
			localSession.MessageCreate(&khl.MessageCreate{
				MessageCreateBase: khl.MessageCreateBase{
					Type:     khl.MessageTypeKMarkdown,
					TargetID: commonChannel,
					Content:  "(met)" + v.Id + "(met) " + randomSentence(resinFull),
				},
			})
			for _, notice := range v.Notices {
				if notice.Type == "countdown" {
					localSession.MessageDelete(notice.Id)
				}
			}
		} else if timeLeft < 1260000 {
			var alreadyCountdown bool = false
			for _, b := range v.Notices {
				if b.Type == "countdown" {
					alreadyCountdown = true
				}
			}
			if !alreadyCountdown {
				card := kcard.KHLCard{}
				card.Init()
				card.Card.Theme = "warning"
				card.AddModule_markdown("(met)" + v.Id + "(met)" + " YUI戳了戳你。提醒你距离你的树脂溢出还有:")
				card.AddModule_divider()
				card.AddModule(
					kcard.KModule{
						Type:    "countdown",
						Mode:    "hour",
						EndTime: v.EndTime,
					},
				)
				msg, _ := sendKCard(commonChannel, card.String())
				v.Notices = append(v.Notices, resinNotice{msg.MsgID, "countdown"})
			}
			resinRecordv2Cache[validIdx] = v
			validIdx++
		} else {
			resinRecordv2Cache[validIdx] = v
			validIdx++
		}
	}
	resinRecordv2Cache = resinRecordv2Cache[:validIdx]
	resinRecordSaveLocal_v2()
}
