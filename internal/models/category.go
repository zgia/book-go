package models

import (
	"zgia.net/book/internal/db"
)

// 接口返回
type CategoryResult struct {
	Id        int64  `json:"id"`
	Parentid  int64  `json:"parentid"`
	Title     string `json:"title"`
	BookCount int64  `json:"bookcount"`
	IsHidden  int64  `json:"ishidden"`
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
			Parentid:  v.Parentid,
			Title:     v.Title,
			BookCount: count[v.Id],
			IsHidden:  v.IsHidden,
		}
	}

	return vr, nil
}
