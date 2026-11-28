package db

import (
	"fmt"
	"strings"
	"time"

	"zgia.net/book/internal/conf"
)

type Author struct {
	Id         int64 `xorm:"pk autoincr"`
	Name       string
	FormerName string
	CreatedAt  time.Time `xorm:"created notnull"`
	UpdatedAt  time.Time `xorm:"updated notnull"`
	DeletedAt  time.Time `xorm:"deleted null"`
}

func TableName() string {
	return "author"
}

func (b *Author) String() string {
	return fmt.Sprintf("AuthorAuthor: %v, id: %d, former name: %v", b.Name, b.Id, b.FormerName)
}

func CheckAuthor(name, formerName string) (int64, error) {
	if name == "" {
		return 0, nil
	}

	var ids []int64
	var err error

	if conf.UseMySQL {
		// FIND_IN_SET 精确匹配逗号分隔的曾用名列表。
		err = x.Table(TableName()).Cols("id").Where("name=? OR FIND_IN_SET(?, former_name)", name, name).Limit(1).Find(&ids)
	} else {
		// SQLite 没有 FIND_IN_SET：把列表包上逗号后用 LIKE 做整词匹配。
		// 转义 % 和 _，避免作者名里的通配符被当成 LIKE 模式。
		escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(name)
		err = x.Table(TableName()).Cols("id").Where("name=? OR (',' || former_name || ',') LIKE ? ESCAPE '\\'", name, "%,"+escaped+",%").Limit(1).Find(&ids)
	}
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
	_, err = x.Table(TableName()).Insert(auth)
	if err != nil {
		return 0, err
	}

	return auth.Id, nil

}

// CountAuthors returns number of authors.
func CountAuthors(words string) (int64, error) {
	words = "%" + words + "%"

	return x.Table(TableName()).Where("name LIKE ? OR former_name LIKE ?", words, words).Count(new(Author))
}

// ListAuthors returns number of authors in given page.
func QueryAuthors(page int, words string) ([]*Author, error) {
	pageSize := conf.PageSize(0)
	words = "%" + words + "%"
	authors := make([]*Author, 0, pageSize)

	return authors, x.Table(TableName()).Where("name LIKE ? OR former_name LIKE ?", words, words).Desc("updated_at").Desc("id").Limit(pageSize, (page-1)*pageSize).Find(&authors)
}

// QueryAuthor gets a author info
func QueryAuthor(authorid int64) (*Author, error) {
	author := &Author{
		Id: authorid,
	}
	has, err := x.Table(TableName()).Get(author)

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
		_, err = sess.Table(TableName()).Insert(author)
		if err != nil {
			return 0, err
		}

		authorid = author.Id
	} else {
		if _, err = sess.Table(TableName()).ID(authorid).Cols("name", "former_name").Update(author); err != nil {
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
	return x.Table(TableName()).Delete(Author{Id: authorid})
}
