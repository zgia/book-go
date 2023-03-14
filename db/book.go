package db

import (
	"fmt"

	"zgia.net/book/conf"
)

type Book struct {
	Id         int64 `xorm:"pk autoincr"`
	Categoryid int64 `xorm:"notnull default 0"`
	Title      string
	Author     string
	Summary    string
	Source     string
	Createdat  int64 `xorm:"created notnull default 0"`
	Updatedat  int64 `xorm:"updated notnull default 0"`
	Deletedat  int64 `xorm:"deleted notnull default 0"`
}

func (b *Book) String() string {
	return fmt.Sprintf("Book: %v(%v), %v", b.Title, b.Author, b.Summary)
}

// CountBooks returns number of books.
func CountBooks(title string) int64 {
	count, _ := x.Where("title LIKE ?", "%"+title+"%").Count(new(Book))

	return count
}

// ListBooks returns number of books in given page.
func QueryBooks(page int, title string) ([]*Book, error) {
	pageSize := conf.PageSize(20)
	books := make([]*Book, 0, pageSize)

	return books, x.Limit(pageSize, (page-1)*pageSize).Where("deletedat=?", 0).And("title LIKE ?", "%"+title+"%").Desc("id").Find(&books)
}

const CHAPTERS_SQL = "SELECT txt AS content,cha.title,vol.title AS volTitle " +
	"FROM chapter AS cha " +
	"INNER JOIN content AS content ON cha.id=content.chapterid " +
	"INNER JOIN volume AS vol ON vol.id=cha.volumeid " +
	"WHERE cha.bookid=? AND cha.deletedat=0 ORDER BY cha.id"

func QueryAllChapters(bookid int64) []map[string]string {

	var results []map[string]string

	if err := x.SQL(CHAPTERS_SQL, bookid).Find(&results); err != nil {
		panic(fmt.Sprintf("Cannot query book(%d) chapters, %s", bookid, err.Error()))
	}

	return results
}

// QueryBook gets a book info
func QueryBook(bookid int64) (*Book, error) {
	book := &Book{
		Id: bookid,
	}
	has, err := x.Where("deletedat=?", 0).Get(book)

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
		_, err = sess.Insert(book)
		if err != nil {
			return 0, err
		}

		bookid = book.Id
	} else {
		if _, err = sess.ID(bookid).Cols("title", "author", "summary", "source").Update(book); err != nil {
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
	return x.ID(bookid).Delete(new(Book))
}
