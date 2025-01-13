package models

import (
	"zgia.net/book/internal/db"
)

// 接口返回
type AuthorResult struct {
	Id         int64    `json:"id"`
	Name       string   `json:"name"`
	FormerName string   `json:"former_name"`
	Books      []string `json:"books"`
}

func ListAuthors(page int, words string) (map[string]any, error) {
	books, err := db.QueryAuthors(page, words)

	if err != nil {
		return nil, err
	}

	bs := make([]*AuthorResult, len(books))

	for i, v := range books {
		bs[i] = GetAuthor(v)
	}

	total, _ := db.CountAuthors(words)
	data := map[string]any{
		"total": total,
		"items": bs,
	}

	return data, nil
}

func GetAuthor(author *db.Author) *AuthorResult {

	return &AuthorResult{
		Id:         author.Id,
		Name:       author.Name,
		FormerName: author.FormerName,
	}
}
