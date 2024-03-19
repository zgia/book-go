package models

import (
	"zgia.net/book/internal/db"

	log "zgia.net/book/internal/logger"
	"zgia.net/book/internal/util"
)

// 接口返回
type BookResult struct {
	Id            int64  `json:"id"`
	Categoryid    int64  `json:"categoryid"`
	Title         string `json:"title"`
	Author        string `json:"author"`
	Summary       string `json:"summary"`
	Source        string `json:"source"`
	Cover         string `json:"cover"`
	Wordcount     int64  `json:"wordcount"`
	Done          int64  `json:"done"`
	LatestChapter string `json:"latestChapter"`
}

func SearchBooks(words string, bookid int64) []map[string]string {
	log.Infof("SearchBooks: %s", words)
	chapters := db.QueryBooksByKeywords(words, bookid)

	for _, v := range chapters {
		v["content"] = util.SearchResult(v["content"], words, 40)

		log.Infof("SearchBooks: %#v", v)
	}

	return chapters
}

func ListBooks(page int, title, searchMode string) (map[string]any, error) {
	books, err := db.QueryBooks(page, title, searchMode)

	if err != nil {
		return nil, err
	}

	bs := make([]*BookResult, len(books))

	bookids := make([]int64, len(books))
	for _, v := range books {
		if v.Done == 0 {
			bookids = append(bookids, v.Id)
		}
	}

	latestChapters := db.QueryLatestChapters(bookids)
	log.Infof("latestChapters: %#v", latestChapters)

	for i, v := range books {
		bs[i] = GetBook(v)
		bs[i].LatestChapter = latestChapters[v.Id]
	}

	total, _ := db.CountBooks(title, searchMode)
	data := map[string]any{
		"total": total,
		"items": bs,
	}

	return data, nil
}

func GetBook(book *db.Book) *BookResult {

	return &BookResult{
		Id:         book.Id,
		Categoryid: book.Categoryid,
		Title:      book.Title,
		Author:     book.Author,
		Summary:    book.Summary,
		Source:     book.Source,
		Cover:      book.Cover,
		Wordcount:  book.Wordcount,
		Done:       book.Done,
	}
}
