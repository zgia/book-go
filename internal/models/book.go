package models

import (
	"zgia.net/book/internal/db"

	log "zgia.net/book/internal/logger"
	"zgia.net/book/internal/util"
)

// 接口返回
type BookResult struct {
	Id               int64  `json:"id"`
	Categoryid       int64  `json:"categoryid"`
	Title            string `json:"title"`
	Author           string `json:"author"`
	AuthorFormerName string `json:"author_former_name"`
	Alias            string `json:"alias"`
	Summary          string `json:"summary"`
	Source           string `json:"source"`
	Cover            string `json:"cover"`
	Wordcount        int64  `json:"wordcount"`
	Isfinished       int64  `json:"isfinished"`
	Latest           string `json:"latest"`
	Rate             int64  `json:"rate"`
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

func ListBooks(page int, words, searchMode, orderby, direction, rate string) (map[string]any, error) {
	books, err := db.QueryBooks(page, words, searchMode, orderby, direction, rate)

	if err != nil {
		return nil, err
	}

	bs := make([]*BookResult, len(books))

	for i, v := range books {
		bs[i] = GetBook(v)
	}

	total, _ := db.CountBooks(words, searchMode, rate)
	data := map[string]any{
		"total": total,
		"items": bs,
	}

	return data, nil
}

func GetBooksSize() map[string]string {
	return db.QueryBooksSize()
}

func GetBook(book *db.Book) *BookResult {

	return &BookResult{
		Id:               book.Id,
		Categoryid:       book.Categoryid,
		Title:            book.Title,
		Author:           book.Author,
		AuthorFormerName: book.AuthorFormerName,
		Alias:            book.Alias,
		Summary:          book.Summary,
		Source:           book.Source,
		Cover:            book.Cover,
		Latest:           book.Latest,
		Wordcount:        book.Wordcount,
		Isfinished:       book.Isfinished,
		Rate:             book.Rate,
	}
}
