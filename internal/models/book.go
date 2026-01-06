package models

import (
	"fmt"
	"os"

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

// 搜索返回
type SearchResult struct {
	Content   string `json:"content"`
	ChaId     string `json:"chaId"`
	ChaTitle  string `json:"chaTitle"`
	VolTitle  string `json:"volTitle"`
	VolId     string `json:"volId"`
	BookId    string `json:"bookId"`
	BookTitle string `json:"bookTitle"`
	Author    string `json:"author"`
}

// 分页书籍返回
type PagedBooksResult struct {
	Total int64         `json:"total"`
	Items []*BookResult `json:"items"`
}

func SearchBooks(words string, bookid int64) ([]SearchResult, error) {
	log.Infof("SearchBooks: %s", words)
	chapters, err := db.QueryBooksByKeywords(words, bookid)
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(chapters))
	for _, v := range chapters {
		content := util.SearchResult(v["content"], words, 40)
		result := SearchResult{
			Content:   content,
			ChaId:     v["chaId"],
			ChaTitle:  v["chaTitle"],
			VolTitle:  v["volTitle"],
			VolId:     v["volId"],
			BookId:    v["bookId"],
			BookTitle: v["bookTitle"],
			Author:    v["author"],
		}
		results = append(results, result)
		log.Infof("SearchBooks: %#v", result)
	}

	return results, nil
}

func ListBooks(page int, words, searchMode, orderby, direction, rate string) (*PagedBooksResult, error) {
	books, err := db.QueryBooks(page, words, searchMode, orderby, direction, rate)

	if err != nil {
		return nil, err
	}

	bs := make([]*BookResult, len(books))

	for i, v := range books {
		bs[i] = GetBook(v)
	}

	total, err := db.CountBooks(words, searchMode, rate)
	if err != nil {
		return nil, err
	}

	result := &PagedBooksResult{
		Total: total,
		Items: bs,
	}

	return result, nil
}

func GetBooksSize() (map[string]string, error) {
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

// SaveAllBooksToTxt saves all books to text files
func SaveAllBooksToTxt(bookId int64, savedPath string) {
	ids, err := db.QueryAllBookIds()
	if err != nil {
		log.Errorf("%v", err)
		return
	}

	i := 0
	for _, id := range ids {
		if id < bookId {
			continue
		}

		book, err := db.QueryBook(id)
		if err != nil {
			log.Errorf("%v", err)
			continue
		}
		if book == nil {
			log.Errorf("book %d is not exist", id)
			continue
		}

		i++
		chapters, err := db.QueryAllChapters(id)
		if err != nil {
			log.Errorf("QueryAllChapters(%d): %v", id, err)
			continue
		}
		fpath := WriteToFile(book, chapters)

		srcInfo, _ := os.Lstat(fpath)
		util.MoveFile(fpath, fmt.Sprintf("%s/%d-%s", savedPath, id, srcInfo.Name()))

		log.Infof("book %d save to %s...", id, fpath)
	}

	log.Infof("%d books saved", i)
}
