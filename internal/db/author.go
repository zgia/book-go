package db

import (
	"fmt"
	"time"

	"zgia.net/book/internal/conf"
)

type Author struct {
	Id         int64 `xorm:"pk autoincr"`
	Name       string
	FormerName string
	Createdat  int64 `xorm:"created notnull default 0"`
	Updatedat  int64 `xorm:"updated notnull default 0"`
	Deletedat  int64 `xorm:"deleted notnull default 0"`
}

func TableName() string {
	return "book_author"
}

func (b *Author) String() string {
	return fmt.Sprintf("AuthorAuthor: %v, id: %d, former name: %v", b.Name, b.Id, b.FormerName)
}

func CheckAuthor(name, formerName string) (int64, error) {
	if name == "" {
		return 0, nil
	}

	var ids []int64
	err := x.Table("book_author").Cols("id").Where("name=? OR FIND_IN_SET(?, former_name)", name, name).Limit(1).Find(&ids)
	if err != nil {
		return 0, err
	}

	if len(ids) > 0 {
		return ids[0], nil
	}

	auth := &Author{
		Name:       name,
		FormerName: formerName,
	}
	_, err = x.Table("book_author").Insert(auth)
	if err != nil {
		return 0, err
	}

	return auth.Id, nil

}

// CountAuthors returns number of authors.
func CountAuthors(words string) (int64, error) {
	words = "%" + words + "%"

	return x.Table("book_author").Where("name LIKE ? OR former_name LIKE ?", words, words).Count(new(Author))
}

// ListAuthors returns number of authors in given page.
func QueryAuthors(page int, words string) ([]*Author, error) {
	pageSize := conf.PageSize(0)
	words = "%" + words + "%"
	authors := make([]*Author, 0, pageSize)

	return authors, x.Table("book_author").Where("name LIKE ? OR former_name LIKE ?", words, words).Desc("updatedat").Desc("id").Limit(pageSize, (page-1)*pageSize).Find(&authors)
}

// QueryAuthor gets a author info
func QueryAuthor(authorid int64) (*Author, error) {
	author := &Author{
		Id: authorid,
	}
	has, err := x.Table("book_author").Get(author)

	if err != nil {
		return nil, err
	}

	if !has {
		return nil, nil
	}

	return author, nil
}

// UpdateAuthor updates/creates a author
func UpdateAuthor(author *Author, authorid int64) (int64, error) {
	var err error

	sess := x.NewSession()
	defer sess.Close()

	if err = sess.Begin(); err != nil {
		return 0, err
	}

	if authorid == 0 {
		author.Createdat = time.Now().Unix()
		author.Updatedat = author.Createdat
		_, err = sess.Table("book_author").Insert(author)
		if err != nil {
			return 0, err
		}

		authorid = author.Id
	} else {
		if _, err = sess.Table("book_author").ID(authorid).Cols("name", "former_name").Update(author); err != nil {
			return 0, err
		}
	}

	if err = sess.Commit(); err != nil {
		return 0, err
	}

	return authorid, nil

}

// DeleteAuthor deletes a author
func DeleteAuthor(authorid int64) (int64, error) {
	return x.Table("book_author").Delete(Author{Id: authorid})
}
