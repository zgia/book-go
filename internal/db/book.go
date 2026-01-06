package db

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"xorm.io/builder"
	"xorm.io/xorm"

	"zgia.net/book/internal/conf"
	log "zgia.net/book/internal/logger"
	"zgia.net/book/internal/util"
)

type Book struct {
	Id               int64 `xorm:"pk autoincr"`
	Categoryid       int64 `xorm:"notnull default 0"`
	Title            string
	Alias            string
	Author           string `xorm:"<-"`
	Authorid         int64
	AuthorFormerName string `xorm:"<-"`
	Summary          string
	Source           string
	Cover            string
	Latest           string
	Rate             int64 `xorm:"notnull default 0"`
	Wordcount        int64 `xorm:"notnull default 0"`
	Isfinished       int64 `xorm:"notnull default 0"`
	Createdat        int64 `xorm:"created notnull default 0"`
	Updatedat        int64 `xorm:"updated notnull default 0"`
	Deletedat        int64 `xorm:"deleted notnull default 0"`
}

func (b *Book) String() string {
	return fmt.Sprintf("Book: %s(%s-%d), %s", b.Title, b.Author, b.Authorid, b.Summary)
}

func condition(session *xorm.Session, words, searchMode, rate string) {

	if words != "" {
		if searchMode == "categ" {
			session.Where("book.categoryid=?", words)
		} else if searchMode == "author" {
			session.Where("au.name LIKE ? OR au.former_name LIKE ?", "%"+words+"%", "%"+words+"%")
		} else {
			session.Where("title LIKE ? OR alias LIKE ?", "%"+words+"%", "%"+words+"%")
		}
	}

	rates := util.ParseRates(rate)
	if len(rates) > 0 {
		session.In("rate", rates)
	}
}

// CountBooks returns number of books.
func CountBooks(words, searchMode, rate string) (int64, error) {
	session := x.Table("book").Join("LEFT", "author AS au", "book.authorid = au.id")
	condition(session, words, searchMode, rate)

	return session.Count(new(Book))

}

// ListBooks returns number of books in given page.
func QueryBooks(page int, words, searchMode, orderby, direction, rate string) ([]*Book, error) {
	// Validate page number
	if page < 1 {
		page = 1
	}
	pageSize := conf.PageSize(0)
	books := make([]*Book, 0, pageSize)

	session := x.Table("book").Join("LEFT", "author AS au", "book.authorid = au.id").Select("book.*,au.name AS author,au.former_name AS author_former_name")
	condition(session, words, searchMode, rate)

	// 排序
	orders := map[string]int{"id": 1, "wordcount": 1, "updatedat": 1, "rate": 1}
	if _, ok := orders[orderby]; !ok {
		orderby = "updatedat"
	}
	orderby = fmt.Sprintf("book.%s", orderby)

	if direction != "ascending" {
		session.Desc(orderby)
	} else {
		session.Asc(orderby)
	}

	return books, session.Limit(pageSize, (page-1)*pageSize).Find(&books)
}

// QueryBooksByAuthorIds returns books by author IDs.
func QueryBooksByAuthorIds(authorIds []int64) ([]*Book, error) {
	books := make([]*Book, 0, len(authorIds))

	return books, x.Table("book").Select("id, title,authorid").In("authorid", authorIds).Find(&books)
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
func QueryLatestChapters(bookids []int64) (map[int64]string, error) {
	var results []map[string]string

	subQuery := builder.Select("MAX(id) AS id").
		From("chapter").
		Where(builder.In("bookid", bookids)).
		GroupBy("bookid")

	sql, args, err := builder.Select("bookid, title").
		From("chapter").
		Where(builder.In("id", subQuery)).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("cannot build SQL for latest chapters: %w", err)
	}

	if err := x.SQL(sql, args...).Find(&results); err != nil {
		return nil, fmt.Errorf("cannot query books latest chapters, sql: %s, error: %w", sql, err)
	}

	chaps := make(map[int64]string, len(results))
	for _, v := range results {
		chaps[util.ParamInt64(v["bookid"])] = v["title"]
	}

	return chaps, nil
}

func QueryVolumeChapters(bookid, volumeid int64) ([]map[string]string, error) {
	volume_chapters_sql := "SELECT txt AS content,cha.title " +
		"FROM chapter AS cha " +
		"INNER JOIN content AS content ON cha.id=content.chapterid " +
		"WHERE cha.bookid=? AND cha.volumeid=? AND cha.deletedat=0 ORDER BY cha.id"

	var results []map[string]string

	if err := x.SQL(volume_chapters_sql, bookid, volumeid).Find(&results); err != nil {
		return nil, fmt.Errorf("cannot query book(%d)-vol(%d) chapters: %w", bookid, volumeid, err)
	}

	return results, nil
}

func QueryAllChapters(bookid int64) ([]map[string]string, error) {
	chapters_sql := "SELECT txt AS content,cha.title,vol.title AS volTitle " +
		"FROM chapter AS cha " +
		"INNER JOIN content AS content ON cha.id=content.chapterid " +
		"INNER JOIN volume AS vol ON vol.id=cha.volumeid " +
		"WHERE cha.bookid=? AND cha.deletedat=0 ORDER BY cha.id"

	var results []map[string]string

	if err := x.SQL(chapters_sql, bookid).Find(&results); err != nil {
		return nil, fmt.Errorf("cannot query book(%d) chapters: %w", bookid, err)
	}

	return results, nil
}

func QueryBooksByKeywords(words string, bookid int64) ([]map[string]string, error) {
	search_books_sql := "SELECT txt AS content,cha.id AS chaId,cha.title AS chaTitle,vol.title AS volTitle " +
		",volumeid AS volId,book.id AS bookId,book.title AS bookTitle,au.name AS author " +
		"FROM chapter AS cha " +
		"INNER JOIN content AS content ON cha.id=content.chapterid " +
		"INNER JOIN volume AS vol ON vol.id=cha.volumeid " +
		"INNER JOIN book AS book ON cha.bookid=book.id " +
		"LEFT JOIN author AS au ON au.id=book.authorid " +
		"WHERE _BOOKID_CONDITION_ content.txt LIKE ? AND cha.deletedat=0 ORDER BY cha.id LIMIT 100"

	var results []map[string]string

	var bookidCond = ""
	var args []any
	if bookid != 0 {
		bookidCond = " cha.bookid = ? AND "
		args = append(args, bookid)
	}
	args = append(args, "%"+words+"%")
	search_books_sql = strings.Replace(search_books_sql, "_BOOKID_CONDITION_", bookidCond, 1)

	if err := x.SQL(search_books_sql, args...).Find(&results); err != nil {
		return nil, fmt.Errorf("cannot query book(%d) by keywords(%s) chapters: %w", bookid, words, err)
	}

	return results, nil
}

func QueryBooksSize() (map[string]string, error) {
	if conf.UseMySQL {
		sql := "SELECT CEILING(sum(data_length)/1024/1024) AS data_size, CEILING(sum(index_length)/1024/1024) AS index_size FROM information_schema.tables WHERE TABLE_SCHEMA = ?"
		var results []map[string]string
		if err := x.SQL(sql, conf.Database.Name).Find(&results); err != nil {
			return nil, fmt.Errorf("cannot query books size: %w", err)
		}
		if len(results) == 0 {
			return map[string]string{"data_size": "0", "index_size": "0"}, nil
		}
		return results[0], nil
	} else if conf.UseSQLite {
		// Get page count and page size via PRAGMA
		var pageSize, pageCount int64
		// Query page_size
		var sizeResults []map[string]string
		if err := x.SQL("PRAGMA page_size;").Find(&sizeResults); err != nil {
			return nil, fmt.Errorf("cannot query SQLite page_size: %w", err)
		}
		if len(sizeResults) > 0 {
			for _, v := range sizeResults[0] {
				pageSize, _ = strconv.ParseInt(v, 10, 64)
				break
			}
		}
		// Query page_count
		var countResults []map[string]string
		if err := x.SQL("PRAGMA page_count;").Find(&countResults); err != nil {
			return nil, fmt.Errorf("cannot query SQLite page_count: %w", err)
		}
		if len(countResults) > 0 {
			for _, v := range countResults[0] {
				pageCount, _ = strconv.ParseInt(v, 10, 64)
				break
			}
		}
		// Calculate total size in bytes
		totalBytes := pageSize * pageCount
		// Convert to MB (ceiling)
		totalMB := (totalBytes + 1024*1024 - 1) / (1024 * 1024)
		if totalMB == 0 {
			totalMB = 1 // at least 1 MB
		}
		return map[string]string{
			"data_size":  strconv.FormatInt(totalMB, 10),
			"index_size": "0",
		}, nil
	}
	return nil, fmt.Errorf("unknown database type")
}

// QueryBook gets a book info
func QueryBook(bookid int64) (*Book, error) {
	book := &Book{
		Id: bookid,
	}
	has, err := x.Table("book").Join("LEFT", "author AS au", "book.authorid = au.id").Select("book.*,au.name AS author,au.former_name AS author_former_name").Get(book)

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

	// 处理作者
	book.Authorid, err = CheckAuthor(book.Author, "")
	log.Infof("%v, %v", book, err)

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
		if _, err = sess.ID(bookid).Cols("categoryid", "title", "authorid", "alias", "summary", "source", "wordcount", "isfinished", "rate").Update(book); err != nil {
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
