package models

import (
	"zgia.net/book/internal/db"
)

// 接口返回
type CategoryResult struct {
	Id        int64  `json:"id"`
	ParentID  int64  `json:"parent_id"`
	Title     string `json:"title"`
	BookCount int64  `json:"book_count"`
	IsHidden  int64  `json:"is_hidden"`
}

func ListCategories() ([]*CategoryResult, error) {

	categories, err := db.QueryCategories()

	if err != nil {
		return nil, err
	}

	count, err := db.QueryCountsByCategory()
	if err != nil {
		return nil, err
	}

	vr := make([]*CategoryResult, len(categories))

	for i, v := range categories {
		vr[i] = &CategoryResult{
			Id:        v.Id,
			ParentID:  v.ParentID,
			Title:     v.Title,
			BookCount: count[v.Id],
			IsHidden:  v.IsHidden,
		}
	}

	return vr, nil
}
