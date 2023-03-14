package db

import (
	"fmt"
)

type Chapter struct {
	Id        int64 `xorm:"pk autoincr"`
	Bookid    int64 `xorm:"notnull default 0"`
	Volumeid  int64 `xorm:"notnull default 0"`
	Title     string
	Createdat int64 `xorm:"created notnull default 0"`
	Updatedat int64 `xorm:"updated notnull default 0"`
	Deletedat int64 `xorm:"deleted notnull default 0"`
}

type Content struct {
	Chapterid int64 `xorm:"pk autoincr"`
	Txt       string
}

func (b *Chapter) String() string {
	return fmt.Sprintf("Chapter: %v, bookid: %d, volumeid: %d", b.Title, b.Bookid, b.Volumeid)
}

// CountChapters returns number of chapters.
func CountChapters() int64 {
	count, _ := x.Count(new(Chapter))
	return count
}

// ListChapters returns number of chapters in given page.
func QueryChapters(bookid int64) ([]*Chapter, error) {
	var chapters []*Chapter

	return chapters, x.Where("bookid=?", bookid).And("deletedat=?", 0).Asc("id").Find(&chapters)
}

// QueryChapter gets a chapter
func QueryChapter(chapterid int64, next int) (map[string]any, error) {
	chapter := &Chapter{
		Id: chapterid,
	}

	has, err := x.Where("deletedat=?", 0).Get(chapter)

	if err != nil {
		return nil, err
	}

	if !has {
		return nil, nil
	}

	current := map[string]any{
		"id":    chapter.Id,
		"title": chapter.Title,
	}

	// 卷
	if chapter.Volumeid != 0 {
		volume := &Volume{Id: chapter.Volumeid}
		if has, err = x.Get(volume); has && err == nil {
			current["volumeTitle"] = volume.Title
			current["volumeid"] = volume.Id
		}
	}

	// 正文
	content := &Content{Chapterid: chapter.Id}
	if has, err = x.Get(content); has && err == nil {
		current["content"] = content.Txt
	}

	data := map[string]any{}

	data["current"] = current

	// 前一篇，后一篇，用于阅读时翻页
	if next != 0 {

		// 下一篇
		nextc := &Chapter{}

		if has, err = x.Where("id>?", chapterid).
			And("bookid=?", chapter.Bookid).
			And("volumeid=?", chapter.Volumeid).
			And("deletedat=?", 0).
			Asc("id").
			Get(nextc); has && err == nil {

			data["next"] = map[string]any{"id": nextc.Id, "title": nextc.Title}
		}

		beforec := &Chapter{}
		if has, err = x.Where("id<?", chapterid).
			And("bookid=?", chapter.Bookid).
			And("volumeid=?", chapter.Volumeid).
			And("deletedat=?", 0).
			Desc("id").
			Get(beforec); has && err == nil {

			data["before"] = map[string]any{"id": beforec.Id, "title": beforec.Title}
		}
	}

	return data, nil
}

// UpdateChapter updates/creates a chapter
func UpdateChapter(chapterid int64, chapter *Chapter, content *Content) (id int64, err error) {

	sess := x.NewSession()
	defer sess.Close()

	if err = sess.Begin(); err != nil {
		return 0, err
	}

	if chapterid == 0 {
		_, err = sess.Insert(chapter)
		if err != nil {
			return 0, err
		}

		content.Chapterid = chapter.Id
		_, err = sess.Insert(content)
		if err != nil {
			return 0, err
		}

		chapterid = chapter.Id
	} else {
		if _, err = sess.ID(chapterid).Update(chapter); err != nil {
			return 0, err
		}

		_, err = sess.Update(content, &Content{Chapterid: chapterid})
		if err != nil {
			return 0, err
		}
	}

	if err = sess.Commit(); err != nil {
		return 0, err
	}

	return chapterid, nil
}

// DeleteChapter deletes a Chapter
func DeleteChapter(chapterid int64) (int64, error) {
	return x.ID(chapterid).Delete(new(Chapter))
}
