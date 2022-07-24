package main

import (
	"fmt"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/lithammer/fuzzysearch/fuzzy"
	scribble "github.com/nanobox-io/golang-scribble"
)

func TestDownload(t *testing.T) {
	var sheetsCache []sheetUpload
	db, _ := scribble.New("./database", nil)
	db.Read("db", "sheets", &sheetsCache)
	for i, v := range sheetsCache {
		if !v.IsDownloaded {
			fmt.Println("Downloading", i, v.Name, v.Url)
			filePath := "./sheets/" + randToken(16) + v.Type
			downloadFileTo(v.Url, filePath)
			sheetsCache[i].FilePath = filePath
			sheetsCache[i].IsDownloaded = true
		}
	}
	if err := db.Write("db", "sheets", sheetsCache); err != nil {
		fmt.Println("Error", err)
	}
}

func TestDuplicate(t *testing.T) {
	var sheetsCache []sheetUpload
	db, _ := scribble.New("./database", nil)
	db.Read("db", "sheets", &sheetsCache)
	base := getRawSheetClean("./sheets/1ecdcf84765f4e5d3ceefc1c02181f57.txt")
	// fmt.Println(base)
	for _, v := range sheetsCache {
		if v.IsDownloaded && v.Type == ".txt" {
			strV := getRawSheetClean(v.FilePath)
			similarity := strutil.Similarity(strV, base, metrics.NewSorensenDice())
			fmt.Printf("相似度 %.1f%% [%s] \n", similarity*100, v.Name)
		}
	}
}

// go test -v -timeout 300s -run ^TestDuplicateAll$ github.com/Nigh/YUI-KHL
func TestDuplicateAll(t *testing.T) {
	var sheetsCache []sheetUpload
	var strV, strB string
	db, _ := scribble.New("./database", nil)
	db.Read("db", "sheets", &sheetsCache)
	f := func(s string) {
		fmt.Println(s)
	}
	for i, v := range sheetsCache[:len(sheetsCache)-1] {
		if v.IsDownloaded && !v.Removed {
			<-time.After(time.Millisecond * time.Duration(10))
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
						if isRemoveDupSheetsTest {
							reply = "\n\n谱面A[" + strconv.Itoa(i) + "]:`" + v.Name + "`\n上传者:(met)" + v.UpID + "(met)"
							reply += "\n谱面B[" + strconv.Itoa(realDupIdx) + "]:`" + b.Name + "`\n上传者:(met)" + b.UpID + "(met)"
							reply += "\n**谱面B** 与 **谱面A** 的相似度为" + strconv.FormatFloat(s1*100, 'f', 2, 64) + "%"
						}
						if s1 < 0.99 {
							s2 := strutil.Similarity(strV, strB, metrics.NewLevenshtein())

							if isRemoveDupSheetsTest {
								reply += "\n相似度在 80% 与 99% 之间, 需要进行`Levenshtein`测试"
							}
							if s2 > 0.95 {

								if isRemoveDupSheetsTest {
									reply += "\n`Levenshtein`测试得分为`" + strconv.FormatFloat(s2*100, 'f', 2, 64) + "` 大于等于 95"
									reply += "\n:x: 判定 谱面B 为**重复谱面** :x:\n将予以删除"
								}
								// removeSheetWithIndex(dupIdx)
								// sheetsCache[idx].Removed = true
							} else {

								if isRemoveDupSheetsTest {
									reply += "\n`Levenshtein`测试得分为`" + strconv.FormatFloat(s2*100, 'f', 2, 64) + "` 小于 95"
									reply += "\n:white_check_mark: 判定 谱面B 为**正常谱面** :white_check_mark:\n将予以保留"
								}
							}
						} else {

							if isRemoveDupSheetsTest {
								reply += "\n相似度超过了 99%"
								reply += "\n:x: 判定 谱面B 为**重复谱面** :x:\n将予以删除"
							}

							// removeSheetWithIndex(dupIdx)
							// sheetsCache[idx].Removed = true
						}

						if isRemoveDupSheetsTest {
							f(reply)
						}
						<-time.After(time.Millisecond * time.Duration(100))
					}
				}
			}
		}
	}
}
func TestFuzzySearch(t *testing.T) {
	var sheetsCache []sheetUpload
	var searchSlice []string
	db, _ := scribble.New("./database", nil)
	db.Read("db", "sheets", &sheetsCache)
	for _, v := range sheetsCache {
		if v.Removed {
			searchSlice = append(searchSlice, "")
		} else {
			searchSlice = append(searchSlice, v.Name)
		}
	}
	matches := fuzzy.RankFindNormalizedFold("极乐", searchSlice)
	sort.Sort(matches)
	if len(matches) == 0 {
		fmt.Println("没有找到匹配")
	} else {
		fmt.Println("找到匹配如下：")
		for _, v := range matches {
			fmt.Println(sheetsCache[v.OriginalIndex].Name, sheetsCache[v.OriginalIndex].Url)
		}
	}
}
func TestSheetsRaw(t *testing.T) {
	fileA := getRawSheetClean("./sheets/ad70384859c05a4b13a1d5b2f5bfef5a.txt")
	fileB := getRawSheetClean("./sheets/c37f4192779ec1d96b3b1c5424fce0b7.txt")
	fmt.Println("FileA", fileA)
	fmt.Println("FileB", fileB)
}
