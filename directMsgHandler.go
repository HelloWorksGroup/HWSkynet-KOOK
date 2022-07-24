package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	kcard "local/khlcard"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/lonelyevil/khl"
)

var lastListID string
var sheets []khl.Attachment
var dupsheets []string

func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func listSheetFrom(ID string) int {
	// var bf bytes.Buffer
	ml, err := localSession.MessageList(uploadChannel,
		khl.MessageListWithMsgID(ID),
		khl.MessageListWithFlag(khl.MessageListFlagAfter))
	if err != nil {
		fmt.Println("ERR:", err.Error())
	} else {
		for _, v := range ml {
			lastListID = v.ID
			if v.Attachments != nil {
				b := v.Attachments
				for _, n := range sheets {
					if b.URL == n.URL {
						dupsheets = append(dupsheets, v.ID)
						break
					}
				}
				sheets = append(sheets, *v.Attachments)
				// bf.WriteString(v.ID + "," + b.Name + "," + b.Type + "," + b.FileType + "," + strconv.FormatInt(b.Size, 10) + "," + b.URL + "\r\n")
				sheetAdd(sheetUpload{
					UpID:         v.Author.ID,
					Id:           v.ID,
					Name:         b.Name,
					Url:          b.URL,
					Type:         filepath.Ext(b.Name),
					Size:         b.Size,
					IsDownloaded: false,
				})
			}
		}
		// sl, _ := os.OpenFile("sheetlist.csv", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
		// defer sl.Close()
		// sl.WriteString(bf.String())
		return len(ml)
	}
	return 0
}

func listSheets(ctxCommon *khl.EventDataGeneral, s []string, f func(string) string) {
	if checkFileIsExist("sheetlist.csv") {
		os.Remove("sheetlist.csv")
	}
	lastListID = "4e46946f-26d9-4855-8c14-315c44bf11ed"
	sheets = []khl.Attachment{}
	dupsheets = []string{}
	for {
		n := listSheetFrom(lastListID)
		if n < 50 {
			break
		}
	}
	f("List Complete!!!\nTotal `" + strconv.Itoa(len(sheets)) + "`\nDuplicate `" + strconv.Itoa(len(dupsheets)) + "`")
	return
}

func cleanSheet(in []byte) []byte {
	found := bytes.Index(in, []byte("====="))
	if found > 0 {
		in = in[found:]
	}
	in = bytes.ToUpper(in)
	n := 0
	for _, val := range in {
		if (val >= 'A' && val <= 'Z') || (val >= '0' && val <= '9') {
			in[n] = val
			n++
		}
	}
	in = in[:n]
	return bytes.ToUpper(in)
}

func getRawSheet(path string) string {
	fileBytes, _ := ioutil.ReadFile(path)
	return string(fileBytes)
}
func getRawSheetClean(path string) string {
	fileBytes, _ := ioutil.ReadFile(path)
	return string(cleanSheet(fileBytes))
}
func similarityBetween(a, b string) float64 {
	return strutil.Similarity(a, b, metrics.NewSorensenDice())
}

// 查询索引为idx的谱面是否与索引<idx的谱面有重复
// 如果没有重复则返回-1，否则返回与之重复的索引
func isDupSheet(idx int) (index int, reason string) {
	var strV, strB string
	if idx >= len(sheetsCache) {
		return -2, "out of bound"
	}
	if sheetsCache[idx].Type == ".txt" {
		strB = getRawSheetClean(sheetsCache[idx].FilePath)
	} else {
		strB = getRawSheet(sheetsCache[idx].FilePath)
	}
	for i, v := range sheetsCache[:idx-1] {
		if v.IsDownloaded && !v.Removed && v.Type == sheetsCache[idx].Type {
			if v.Type == ".txt" {
				strV = getRawSheetClean(v.FilePath)
			} else {
				strV = getRawSheet(v.FilePath)
			}
			s1 := similarityBetween(strV, strB)
			if s1 > 0.8 {
				if s1 < 0.99 {
					s2 := strutil.Similarity(strV, strB, metrics.NewLevenshtein())
					if s2 > 0.95 {
						return i, "相似度达到" + strconv.FormatFloat(s2*100, 'f', 2, 64) + "%"
					} else {
						return -1, ""
					}
				} else {
					return i, "相似度超过99%"
				}
			}
		}
	}
	return -1, ""
}

func removedSheetsRecover(ctxCommon *khl.EventDataGeneral, s []string, f func(string) string) {
	for i, v := range sheetsCache {
		if v.Removed == true {
			sheetsCache[i].Removed = false
		}
	}
	if err := db.Write("db", "sheets", sheetsCache); err != nil {
		f("ERROR:" + err.Error())
	}
	f("已恢复")
}

func sheetsRecover(ctxCommon *khl.EventDataGeneral, s []string, f func(string) string) {
	for i, v := range sheetsCache {
		if v.Id == s[1] {
			sheetsCache[i].Removed = false
			break
		}
	}
	if err := db.Write("db", "sheets", sheetsCache); err != nil {
		f("ERROR:" + err.Error())
	}
	f("已恢复")
}

var isRemoveDupSheetsTest bool = true

func removeDupSheets(ctxCommon *khl.EventDataGeneral, s []string, f func(string) string) {
	var strV, strB string
	if !isRemoveDupSheetsTest {
		sendMarkdown(sheetChannel, "**谱面查重系统广播**\n---\n谱面查重全面扫描开始，将即时播报谱面查重结果")
	}
	for i, v := range sheetsCache[:len(sheetsCache)-1] {
		if v.IsDownloaded && !v.Removed {
			<-time.After(time.Millisecond * time.Duration(100))
			if v.Type == ".txt" {
				strV = getRawSheetClean(v.FilePath)
			} else {
				strV = getRawSheet(v.FilePath)
			}
			for dupIdx, b := range sheetsCache[i+1:] {
				realDupIdx := dupIdx + i + 1
				if b.IsDownloaded && !b.Removed && v.Type == b.Type {
					if b.Type == ".txt" {
						strB = getRawSheetClean(b.FilePath)
					} else {
						strB = getRawSheet(b.FilePath)
					}
					s1 := similarityBetween(strV, strB)
					if s1 > 0.8 {
						var reply string
						var card kcard.KHLCard
						if !isRemoveDupSheetsTest {
							card = kcard.KHLCard{}
							card.Init()
							card.AddModule_header("谱面查重系统广播")
							card.AddModule_divider()
							card.AddModule_markdown("谱面A:`" + v.Name + "`\n上传者:(met)" + v.UpID + "(met)")
							card.AddModule_divider()
							card.AddModule_markdown("谱面B:`" + b.Name + "`\n上传者:(met)" + b.UpID + "(met)")
							card.AddModule_divider()
							card.AddModule_markdown("**谱面B** 与 **谱面A** 的相似度为" + strconv.FormatFloat(s1*100, 'f', 2, 64) + "%")
						}
						if isRemoveDupSheetsTest {
							reply = "谱面A:`" + v.Name + "`\n上传者:(met)" + v.UpID + "(met)"
							reply += "\n谱面B:`" + b.Name + "`\n上传者:(met)" + b.UpID + "(met)"
							reply += "\n**谱面B** 与 **谱面A** 的相似度为" + strconv.FormatFloat(s1*100, 'f', 2, 64) + "%"
						}
						if s1 < 0.99 {
							s2 := strutil.Similarity(strV, strB, metrics.NewLevenshtein())
							if !isRemoveDupSheetsTest {
								card.AddModule_markdown("相似度在 80% 与 99% 之间, 需要进行`Levenshtein`测试\n")
							}
							if isRemoveDupSheetsTest {
								reply += "相似度在 80% 与 99% 之间, 需要进行`Levenshtein`测试\n"
							}
							if s2 > 0.95 {
								if !isRemoveDupSheetsTest {
									card.AddModule_markdown("`Levenshtein`测试得分为`" + strconv.FormatFloat(s2*100, 'f', 2, 64) + "` 大于等于 95")
									card.AddModule_divider()
									card.Card.Theme = "danger"
									card.AddModule_markdown(":x: 判定 谱面B 为**重复谱面** :x:\n将予以删除")
								}
								if isRemoveDupSheetsTest {
									reply += "\n`Levenshtein`测试得分为`" + strconv.FormatFloat(s2*100, 'f', 2, 64) + "` 大于等于 95"
									reply += "\n:x: 判定 谱面B 为**重复谱面** :x:\n将予以删除"
								}
								removeSheetWithIndex(realDupIdx)
							} else {
								if !isRemoveDupSheetsTest {
									card.AddModule_markdown("`Levenshtein`测试得分为`" + strconv.FormatFloat(s2*100, 'f', 2, 64) + "` 小于 95")
									card.AddModule_divider()
									card.Card.Theme = "success"
									card.AddModule_markdown(":white_check_mark: 判定 谱面B 为**正常谱面** :white_check_mark:\n将予以保留")
								}
								if isRemoveDupSheetsTest {
									reply += "\n`Levenshtein`测试得分为`" + strconv.FormatFloat(s2*100, 'f', 2, 64) + "` 小于 95"
									reply += "\n:white_check_mark: 判定 谱面B 为**正常谱面** :white_check_mark:\n将予以保留"
								}
							}
						} else {
							if !isRemoveDupSheetsTest {
								card.AddModule_markdown("相似度超过了 99%")
								card.AddModule_divider()
								card.Card.Theme = "danger"
								card.AddModule_markdown(":x: 判定 谱面B 为**重复谱面** :x:\n将予以删除")
							}
							if isRemoveDupSheetsTest {
								reply += "\n相似度超过了 99%"
								reply += "\n:x: 判定 谱面B 为**重复谱面** :x:\n将予以删除"
							}
							removeSheetWithIndex(realDupIdx)
						}
						if !isRemoveDupSheetsTest {
							sendKCard(sheetChannel, card.String())
						}
						if isRemoveDupSheetsTest {
							f(reply)
						}
						<-time.After(time.Millisecond * time.Duration(5000))
					}
				}
			}
		}
	}
	if !isRemoveDupSheetsTest {
		sendMarkdown(sheetChannel, "**谱面查重系统广播**\n---\n谱面查重全面扫描已完成，谱面查重结果见上方播报")
	}
}

func getSheetsInfo(ctxCommon *khl.EventDataGeneral, s []string, f func(string) string) {
	i, _ := strconv.Atoi(s[1])
	if i >= len(sheetsCache) {
		f("`错误` 获取了超出范围的谱面")
	} else {
		v := sheetsCache[i]
		f("" +
			"**上传者:** (met)" + v.UpID + "(met) " + v.UpID + "\n" +
			"**消息ID:** " + v.Id + "\n" +
			"**文件:** [" + v.Name + "](" + v.Url + ")\n" +
			"**类型:** `" + v.Type + "`\n" +
			"**大小:** " + strconv.FormatInt(v.Size, 10) + " bytes")
	}
}
func sheetsInfoPublish(ctxCommon *khl.EventDataGeneral, s []string, f func(string) string) {
	var dmsCount int = 0
	var txtCount int = 0
	var uploaders []struct {
		ID    string
		count int
	}
	var uploadersRanking int = 0
	for _, v := range sheetsCache {
		if v.Removed {
			continue
		}
		if strings.ToLower(v.Type) == ".txt" {
			txtCount += 1
		}
		if strings.ToLower(v.Type) == ".dms" {
			dmsCount += 1
		}
		found := false
		for ii, vv := range uploaders {
			if v.UpID == vv.ID {
				uploaders[ii].count = uploaders[ii].count + 1
				found = true
				break
			}
		}
		if !found {
			uploaders = append(uploaders, struct {
				ID    string
				count int
			}{v.UpID, 1})
		}
	}
	uploadersRanking = len(uploaders)
	if uploadersRanking > 7 {
		uploadersRanking = 7
	}
	sort.SliceStable(uploaders, func(i, j int) bool {
		return uploaders[i].count > uploaders[j].count
	})
	card := kcard.KHLCard{}
	card.Init()
	card.Card.Theme = "success"
	card.AddModule_header("YUI 统计了一下下当前群内的谱面信息哦")
	card.AddModule_divider()
	str := "(chn)" + uploadChannel + "(chn)频道中 一共有谱面 `" + strconv.Itoa(txtCount+dmsCount) + "` 份"
	str += "\n" + "\t其中开源谱面 `" + strconv.Itoa(txtCount) + "` 份，加密谱面 `" + strconv.Itoa(dmsCount) + "` 份"
	str += "\n" + "共有 `" + strconv.Itoa(len(uploaders)) + "` 位吟游诗人贡献了谱面\n其中，排名前 `" + strconv.Itoa(uploadersRanking) + "` 的诗人贡献谱面的数量如下:"
	for i := 0; i < uploadersRanking; i++ {
		str += "\n\t**" + strconv.Itoa(i+1) + ".** (met)" + uploaders[i].ID + "(met) - `" + strconv.Itoa(uploaders[i].count) + "`"
	}
	card.AddModule_markdown(str)
	sendKCard(sheetChannel, card.String())
	return
}

func portMarkdown(ctxCommon *khl.EventDataGeneral, s []string, f func(string) string) {
	sendMarkdown(s[1], s[2])
	return
}

func downloadFileTo(url, filepath string) {
	file, err := os.Create(filepath)
	if err != nil {
		log.Fatal(err)
	}
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	resp, err := client.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	size, err := io.Copy(file, resp.Body)
	fmt.Println("Downloaded", size, "bytes")
}

func randToken(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "noname"
	}
	return hex.EncodeToString(bytes)
}

func sheetsDownload(ctxCommon *khl.EventDataGeneral, s []string, f func(string) string) {
	f("下载开始")
	for i, v := range sheetsCache {
		if !v.IsDownloaded {
			fmt.Println("Downloading", i, v.Name, v.Url)
			sheetsCache[i].FilePath = "./sheets/" + randToken(16) + v.Type
			downloadFileTo(v.Url, sheetsCache[i].FilePath)
			sheetsCache[i].IsDownloaded = true
		}
	}
	f("下载完成")
	if err := db.Write("db", "sheets", sheetsCache); err != nil {
		fmt.Println("Error", err)
	}
}

func removeSheetWithIndex(idx int) error {
	sheetsCache[idx].Removed = true
	fmt.Println("try to remove MSG", sheetsCache[idx].Id)
	return localSession.MessageDelete(sheetsCache[idx].Id)
}

// 将谱面标记为删除状态
func sheetsRemove(ctxCommon *khl.EventDataGeneral, s []string, f func(string) string) {
	for i, v := range sheetsCache {
		if v.Id == s[1] {
			if v.Removed != true {
				err := removeSheetWithIndex(i)
				if err != nil {
					f("ERROR:" + err.Error())
				} else {
					f("谱面成功删除")
				}
				if err := db.Write("db", "sheets", sheetsCache); err != nil {
					f("ERROR:" + err.Error())
				}
			} else {
				f("谱面此前已经删除，尝试再次删除消息")
				err := localSession.MessageDelete(v.Id)
				if err != nil {
					f("ERROR:" + err.Error())
				}
			}
			return
		}
	}
	f("未找到对应记录")
}

// 永久删除谱面信息
// TODO: 新建存档数据库，将信息移入存档
func sheetsRemoveHard(ctxCommon *khl.EventDataGeneral, s []string, f func(string) string) {
	for i, v := range sheetsCache {
		if v.Id == s[1] {
			err := removeSheetWithIndex(i)
			if err != nil {
				f("ERROR:" + err.Error())
			} else {
				f("谱面消息成功删除")
			}
			if err := db.Write("db", "sheets", sheetsCache); err != nil {
				f("ERROR:" + err.Error())
			}
			return
		}
	}
	f("未找到对应记录")
}

var directRules []handlerRule = []handlerRule{
	{`^\s*list all sheets\s*$`, listSheets},
	{`^\s*sheetduplicatetest\s*$`, removeDupSheets},
	{`^\s*sheetinfo\s*(\d+)$`, getSheetsInfo},
	{`^\s*sheetsum\s*$`, func(ctxCommon *khl.EventDataGeneral, s []string, f func(string) string) {
		f("当前共有谱面`" + strconv.Itoa(len(sheetsCache)) + "`份")
		return
	}},
	{`^\s*sheetinfopub\s*$`, sheetsInfoPublish},
	{`^\s*sheetdownload\s*$`, sheetsDownload},
	{`^\s*sheetremove\s*([0-9a-f\-]{16,48})$`, sheetsRemove},
	{`^\s*sheetremovehard\s*([0-9a-f\-]{16,48})$`, sheetsRemoveHard},
	{`^\s*sheetremovedrecoverall\s*$`, removedSheetsRecover},
	{`^\s*sheetrecover\s+([0-9a-f\-]{16,48})$`, sheetsRecover},
	{`^\s*send\s*(\d+),(.*)$`, portMarkdown},
}

func directMessageHandler(ctxCommon *khl.EventDataGeneral) {
	if ctxCommon.AuthorID != masterID {
		return
	}
	reply := func(words string) string {
		resp, _ := sendMarkdownDirect(masterID, words)
		return resp.MsgID
	}
	fmt.Println("Master said: " + ctxCommon.Content)

	for n := range directRules {
		v := &directRules[n]
		r := regexp.MustCompile(v.matcher)
		matchs := r.FindStringSubmatch(ctxCommon.Content)
		if len(matchs) > 0 {
			go v.getter(ctxCommon, matchs, reply)
			return
		}
	}
}
