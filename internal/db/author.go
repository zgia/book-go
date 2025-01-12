package db

import (
	"fmt"
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
	return fmt.Sprintf("BookAuthor: %v, id: %d, former name: %v", b.Name, b.Id, b.FormerName)
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
