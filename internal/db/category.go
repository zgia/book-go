package db

import (
	"fmt"
)

type Category struct {
	Id       int64 `xorm:"pk autoincr"`
	Parentid int64
	Title    string
}

// 接口返回
type CategoryResult struct {
	Id       int64
	Parentid int64
	Title    string
}

func (b *Category) String() string {
	return fmt.Sprintf("Category: %d, parentid: %d, title: %s", b.Id, b.Parentid, b.Title)
}

// QueryCategories returns number of categories in given page.
func QueryCategories() ([]*Category, error) {
	var categories []*Category
	return categories, x.Asc("id").Find(&categories)
}

// UpdateCategory updates/creates a category
func UpdateCategory(cat *Category, catid int64) (int64, error) {
	var err error

	sess := x.NewSession()
	defer sess.Close()

	if err = sess.Begin(); err != nil {
		return 0, err
	}

	if catid == 0 {
		_, err = sess.Insert(cat)
		if err != nil {
			return 0, err
		}

		catid = cat.Id
	} else {
		if _, err = sess.ID(catid).Cols("id", "title").Update(cat); err != nil {
			return 0, err
		}
	}

	if err = sess.Commit(); err != nil {
		return 0, err
	}

	return catid, nil

}

// DeleteCategory deletes a category
func DeleteCategory(catid int64) (int64, error) {
	return x.Delete(Category{Id: catid})
}
