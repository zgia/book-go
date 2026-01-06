package models

import (
	"fmt"

	"zgia.net/book/internal/db"
	"zgia.net/book/internal/util"
)

// 接口返回
type ChapterResult struct {
	Id        int64  `json:"id"`
	Title     string `json:"title"`
	Wordcount int64  `json:"wordcount"`
}

func GetVolumeChapters(book *db.Book, chapters []map[string]string) string {
	content := ""
	for _, v := range chapters {
		// 章节
		content = (fmt.Sprintf("%s%s%s%s%s", content, v["title"], util.Eol(2), v["content"], util.Eol(2)))
	}

	return content
}

// ListChapters gets all chapters
func ListChapters(book *db.Book) (map[string]any, error) {

	chapters, err := db.QueryChapters(book.Id)

	if err != nil {
		return nil, err
	}

	volumes, err := db.QueryVolumes(book.Id)
	if err != nil {
		return nil, err
	}
	if len(volumes) == 0 {
		volumes = make([]*db.Volume, 1)
		volumes[0] = &db.Volume{Id: 0, Title: book.Title, Summary: ""}
	}

	data := make(map[int64][]*ChapterResult)
	for _, v := range chapters {
		data[v.Volumeid] = append(data[v.Volumeid], &ChapterResult{
			Id:        v.Id,
			Title:     v.Title,
			Wordcount: v.Wordcount,
		})
	}

	if len(data) == 0 {
		var empty []*ChapterResult
		for _, v := range volumes {
			data[v.Id] = empty
		}
	}

	vols := make([]*VolumeResult, len(volumes))
	for i, v := range volumes {
		vols[i] = GetVolume(v)
	}

	dt := map[string]any{
		"items":   data,
		"volumes": vols,
		"book":    GetBook(book),
	}
	return dt, nil
}
