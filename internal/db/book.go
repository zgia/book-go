package db

import (
	"fmt"
	"strings"
	"time"

	"xorm.io/builder"

	"zgia.net/book/internal/conf"
	"zgia.net/book/internal/util"
)

type Book struct {
	Id         int64 `xorm:"pk autoincr"`
	Categoryid int64 `xorm:"notnull default 0"`
	Title      string
	Author     string
	Summary    string
	Source     string
	Cover      string
	Wordcount  int64 `xorm:"notnull default 0"`
	Done       int64 `xorm:"notnull default 0"`
	Createdat  int64 `xorm:"created notnull default 0"`
	Updatedat  int64 `xorm:"updated notnull default 0"`
	Deletedat  int64 `xorm:"deleted notnull default 0"`
}

func (b *Book) String() string {
	return fmt.Sprintf("Book: %v(%v), %v", b.Title, b.Author, b.Summary)
}

func condition(title, searchMode string) (string, string) {
	var cond = ""
	if title != "" {
		if searchMode == "categ" {
			cond = "categoryid = ?"
		} else if searchMode == "author" {
			cond = "author LIKE ?"
			title = "%" + title + "%"
		} else {
			cond = "title LIKE ?"
			title = "%" + title + "%"
		}
	}

	return cond, title
}

// CountBooks returns number of books.
func CountBooks(title, searchMode string) (int64, error) {
	cond, title := condition(title, searchMode)

	return x.Where(cond, title).Count(new(Book))

}

// ListBooks returns number of books in given page.
func QueryBooks(page int, title, searchMode string) ([]*Book, error) {
	pageSize := conf.PageSize(0)
	books := make([]*Book, 0, pageSize)
	cond, title := condition(title, searchMode)

	return books, x.Where(cond, title).Desc("updatedat").Desc("id").Limit(pageSize, (page-1)*pageSize).Find(&books)
}

func QueryAllBookIds() ([]int64, error) {
	var ids []int64
	err := x.Table("book").Cols("id").Find(&ids)

	return ids, err
}

type categoryCount struct {
	Total      int64
	Categoryid int64
}

// QueryCountsByCategory returns books count group by categoryid.
func QueryCountsByCategory() (map[int64]int64, error) {

	res := []categoryCount{}

	err := x.Table("book").Select("categoryid, COUNT(*) AS total").Where("deletedat = 0").GroupBy("categoryid").Find(&res)

	count := map[int64]int64{}
	for _, v := range res {
		count[v.Categoryid] = v.Total
	}

	return count, err
}

// QueryLatestChapters returns latest chapters of given books.
func QueryLatestChapters(bookids []int64) map[int64]string {

	var results []map[string]string

	sql, _ := builder.Select("bookid, title").
		From("chapter").
		Where(
			builder.In("id",
				builder.Expr(
					builder.Select("MAX(id) AS id").
						From("chapter").
						Where(builder.In("bookid", bookids)).
						GroupBy("bookid").
						ToBoundSQL()))).
		ToBoundSQL()

	if err := x.SQL(sql).Find(&results); err != nil {
		panic(fmt.Sprintf("Cannot query books latest chapters, sql: %s, error: %s", sql, err.Error()))
	}

	chaps := make(map[int64]string, len(results))
	for _, v := range results {
		chaps[util.ParamInt64(v["bookid"])] = v["title"]
	}

	return chaps
}

func QueryVolumeChapters(bookid, volumeid int64) []map[string]string {
	volume_chapters_sql := "SELECT txt AS content,cha.title " +
		"FROM chapter AS cha " +
		"INNER JOIN content AS content ON cha.id=content.chapterid " +
		"WHERE cha.bookid=? AND cha.volumeid=? AND cha.deletedat=0 ORDER BY cha.id"

	var results []map[string]string

	if err := x.SQL(volume_chapters_sql, bookid, volumeid).Find(&results); err != nil {
		panic(fmt.Sprintf("Cannot query book(%d)-vol(%d) chapters, %s", bookid, volumeid, err.Error()))
	}

	return results
}

func QueryAllChapters(bookid int64) []map[string]string {
	chapters_sql := "SELECT txt AS content,cha.title,vol.title AS volTitle " +
		"FROM chapter AS cha " +
		"INNER JOIN content AS content ON cha.id=content.chapterid " +
		"INNER JOIN volume AS vol ON vol.id=cha.volumeid " +
		"WHERE cha.bookid=? AND cha.deletedat=0 ORDER BY cha.id"

	var results []map[string]string

	if err := x.SQL(chapters_sql, bookid).Find(&results); err != nil {
		panic(fmt.Sprintf("Cannot query book(%d) chapters, %s", bookid, err.Error()))
	}

	return results
}

func QueryBooksByKeywords(words string, bookid int64) []map[string]string {

	search_books_sql := "SELECT txt AS content,cha.id AS chaId,cha.title AS chaTitle,vol.title AS volTitle " +
		",volumeid AS volId,book.id AS bookId,book.title AS bookTitle,author " +
		"FROM chapter AS cha " +
		"INNER JOIN content AS content ON cha.id=content.chapterid " +
		"INNER JOIN volume AS vol ON vol.id=cha.volumeid " +
		"INNER JOIN book AS book ON cha.bookid=book.id " +
		"WHERE BOOKID_CONDITION content.txt LIKE ? AND cha.deletedat=0 ORDER BY cha.id LIMIT 100"

	var results []map[string]string

	var bookidCond = ""
	var args []any
	if bookid != 0 {
		bookidCond = " cha.bookid = ? AND "
		args = append(args, bookid)
	}
	args = append(args, "%"+words+"%")
	search_books_sql = strings.Replace(search_books_sql, "BOOKID_CONDITION", bookidCond, 1)

	if err := x.SQL(search_books_sql, args...).Find(&results); err != nil {
		panic(fmt.Sprintf("Cannot query book(%d) by keywords(%s) chapters, %s", bookid, words, err.Error()))
	}

	return results
}

// QueryBook gets a book info
func QueryBook(bookid int64) (*Book, error) {
	book := &Book{
		Id: bookid,
	}
	has, err := x.Get(book)

	if err != nil {
		return nil, err
	}

	if !has {
		return nil, nil
	}

	return book, nil
}

// UpdateBook updates/creates a book
func UpdateBook(book *Book, bookid int64) (int64, error) {
	var err error

	sess := x.NewSession()
	defer sess.Close()

	if err = sess.Begin(); err != nil {
		return 0, err
	}

	if bookid == 0 {
		book.Createdat = time.Now().Unix()
		book.Updatedat = book.Createdat
		_, err = sess.Insert(book)
		if err != nil {
			return 0, err
		}

		volume := &Volume{
			Title:  book.Title,
			Bookid: book.Id,
		}
		_, err = sess.Insert(volume)
		if err != nil {
			return 0, err
		}

		bookid = book.Id
	} else {
		if _, err = sess.ID(bookid).Cols("categoryid", "title", "author", "summary", "source", "wordcound", "done").Update(book); err != nil {
			return 0, err
		}
	}

	if err = sess.Commit(); err != nil {
		return 0, err
	}

	return bookid, nil

}

// DeleteBook deletes a book
func DeleteBook(bookid int64) (int64, error) {
	return x.Delete(Book{Id: bookid})
}
