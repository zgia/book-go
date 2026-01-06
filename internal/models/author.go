package models

import (
	"zgia.net/book/internal/db"
)

type AuthorBooks struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
}

// 接口返回
type AuthorResult struct {
	Id         int64         `json:"id"`
	Name       string        `json:"name"`
	FormerName string        `json:"former_name"`
	Books      []AuthorBooks `json:"books"`
}

func ListAuthors(page int, words string) (map[string]any, error) {
	authors, err := db.QueryAuthors(page, words)

	if err != nil {
		return nil, err
	}

	// 获取作者写的书
	var aids []int64
	for _, v := range authors {
		aids = append(aids, v.Id)
	}

	grouped := make(map[int64][]AuthorBooks)
	if len(aids) > 0 {
		books, err := db.QueryBooksByAuthorIds(aids)
		if err != nil {
			return nil, err
		}

		for _, book := range books {
			res := AuthorBooks{
				Id:    book.Id,
				Title: book.Title,
			}
			grouped[book.Authorid] = append(grouped[book.Authorid], res)
		}
	}
	// log.Infof("ListAuthors: %#v", grouped)

	// 保留必要元素
	ats := make([]*AuthorResult, len(authors))
	for i, v := range authors {
		ats[i] = GetAuthor(v)
		ats[i].Books = grouped[v.Id]
	}

	total, _ := db.CountAuthors(words)
	data := map[string]any{
		"total": total,
		"items": ats,
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
