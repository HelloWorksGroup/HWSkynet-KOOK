package main

import "fmt"

type sheetUpload struct {
	UpName       string // 不是从社群上传则使用Name
	UpID         string // 否则使用ID
	Id           string // MsgID
	Name         string // FileName
	Url          string
	Type         string // dms/txt
	Size         int64
	IsDownloaded bool   // 是否已下载到本地
	FilePath     string // 本地文件路径
	Removed      bool
}

var sheetsCache []sheetUpload

func sheetsDbSave() {
	if err := db.Write("db", "sheets", sheetsCache); err != nil {
		fmt.Println("Error", err)
	}
}

func sheetsDbLoad() {
	db.Read("db", "sheets", &sheetsCache)
}

func sheetManagerInit() {
	sheetsDbLoad()
}

// ret: -1 is duplicate sheet, else return the index
func sheetAdd(su sheetUpload) int {
	for _, v := range sheetsCache {
		if su.Url == v.Url {
			return -1
		}
	}
	sheetsCache = append(sheetsCache, su)
	return len(sheetsCache) - 1
}
