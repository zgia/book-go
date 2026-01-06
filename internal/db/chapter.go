package db

import (
	"fmt"
	"time"

	log "zgia.net/book/internal/logger"
	"zgia.net/book/internal/util"
)

type Chapter struct {
	Id        int64 `xorm:"pk autoincr"`
	Bookid    int64 `xorm:"notnull default 0"`
	Volumeid  int64 `xorm:"notnull default 0"`
	Wordcount int64 `xorm:"notnull default 0"`
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
func CountChapters() (int64, error) {
	count, err := x.Count(new(Chapter))
	return count, err
}

// ListChapters returns number of chapters in given page.
func QueryChapters(bookid int64) ([]*Chapter, error) {
	var chapters []*Chapter

	return chapters, x.Where("bookid=?", bookid).Asc("id").Find(&chapters)
}

// QueryChapter gets a chapter
func QueryChapter(bookid, chapterid int64, next int) (map[string]any, error) {
	chapter := &Chapter{
		Id:     chapterid,
		Bookid: bookid,
	}

	has, err := x.Get(chapter)

	if err != nil {
		return nil, err
	}

	if !has {
		return nil, nil
	}

	log.Infof("QueryChapter: %#v", chapter)

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

		volumes, _ := QueryVolumes(chapter.Bookid)
		log.Infof("Volumes: %#v", volumes)
		prevVolId, nextVolId := util.GetPrevNextId(volumes, chapter.Volumeid)

		log.Infof("prevVolId= %d, nextVolId= %d", prevVolId, nextVolId)

		// 下一篇
		nextc := &Chapter{}

		// 先检查同一卷内的
		has, err = x.Where("id>?", chapterid).And("bookid=?", chapter.Bookid).And("volumeid=?", chapter.Volumeid).Asc("id").Get(nextc)
		if has && err == nil {
			data["next"] = map[string]any{"id": nextc.Id, "title": nextc.Title}
		} else if nextVolId != 0 {
			has, err = x.Where("bookid=?", chapter.Bookid).And("volumeid=?", nextVolId).Asc("id").Get(nextc)
			if has && err == nil {
				data["next"] = map[string]any{"id": nextc.Id, "title": nextc.Title}
			}
		}

		prevc := &Chapter{}
		// 先检查同一卷内的
		has, err = x.Where("id<?", chapterid).And("bookid=?", chapter.Bookid).And("volumeid=?", chapter.Volumeid).Desc("id").Get(prevc)
		if has && err == nil {
			data["prev"] = map[string]any{"id": prevc.Id, "title": prevc.Title}
		} else if prevVolId != 0 {
			has, err = x.Where("bookid=?", chapter.Bookid).And("volumeid=?", prevVolId).Desc("id").Get(prevc)
			if has && err == nil {
				data["prev"] = map[string]any{"id": prevc.Id, "title": prevc.Title}
			}
		}
	}

	return data, nil
}

// UpdateChapter updates/creates a chapter
func UpdateChapter(chapterid int64, chapter *Chapter, content *Content) (int64, error) {
	var err error

	sess := x.NewSession()
	defer sess.Close()

	if err = sess.Begin(); err != nil {
		return 0, err
	}

	cols := "wordcount,updatedat"
	book := &Book{Updatedat: time.Now().Unix()}

	if chapterid == 0 {
		_, err = sess.Insert(chapter)
		if err != nil {
			sess.Rollback()
			return 0, err
		}

		content.Chapterid = chapter.Id
		_, err = sess.Insert(content)
		if err != nil {
			sess.Rollback()
			return 0, err
		}

		// 最新章节
		book.Latest = chapter.Title
		cols = "latest," + cols

		chapterid = chapter.Id
	} else {
		if _, err = sess.ID(chapterid).Update(chapter); err != nil {
			sess.Rollback()
			return 0, err
		}

		_, err = sess.Update(content, &Content{Chapterid: chapterid})
		if err != nil {
			sess.Rollback()
			return 0, err
		}
	}

	// 更新Book字数
	total, err := sess.Where("bookid=?", chapter.Bookid).SumInt(new(Chapter), "wordcount")
	if err != nil {
		sess.Rollback()
		return 0, err
	}
	book.Wordcount = total

	if _, err = sess.ID(chapter.Bookid).Cols(cols).Update(book); err != nil {
		sess.Rollback()
		return 0, err
	}

	if err = sess.Commit(); err != nil {
		sess.Rollback()
		return 0, err
	}

	return chapterid, nil
}

// DeleteChapter deletes a Chapter
func DeleteChapter(chapterid int64) (int64, error) {
	return x.Delete(Chapter{Id: chapterid})
}
