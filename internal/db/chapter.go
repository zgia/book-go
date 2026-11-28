package db

import (
	"fmt"
	"time"

	log "zgia.net/book/internal/logger"
	"zgia.net/book/internal/util"
)

type Chapter struct {
	Id        int64 `xorm:"pk autoincr"`
	BookID    int64 `xorm:"notnull default 0"`
	VolumeID  int64 `xorm:"notnull default 0"`
	WordCount int64 `xorm:"notnull default 0"`
	Title     string
	CreatedAt time.Time `xorm:"created notnull"`
	UpdatedAt time.Time `xorm:"updated notnull"`
	DeletedAt time.Time `xorm:"deleted null"`
}

type Content struct {
	ChapterID int64 `xorm:"pk autoincr"`
	Txt       string
}

func (b *Chapter) String() string {
	return fmt.Sprintf("Chapter: %v, book_id: %d, volume_id: %d", b.Title, b.BookID, b.VolumeID)
}

// CountChapters returns number of chapters.
func CountChapters() (int64, error) {
	count, err := x.Count(new(Chapter))
	return count, err
}

// ListChapters returns number of chapters in given page.
func QueryChapters(bookid int64) ([]*Chapter, error) {
	var chapters []*Chapter

	return chapters, x.Where("book_id=?", bookid).Asc("id").Find(&chapters)
}

// QueryChapter gets a chapter
func QueryChapter(bookid, chapterid int64, next int) (map[string]any, error) {
	chapter := &Chapter{
		Id:     chapterid,
		BookID: bookid,
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
	if chapter.VolumeID != 0 {
		volume := &Volume{Id: chapter.VolumeID}
		if has, err = x.Get(volume); has && err == nil {
			current["volumeTitle"] = volume.Title
			current["volume_id"] = volume.Id
		}
	}

	// 正文
	content := &Content{ChapterID: chapter.Id}
	if has, err = x.Get(content); has && err == nil {
		current["content"] = content.Txt
	}

	data := map[string]any{}

	data["current"] = current

	// 前一篇，后一篇，用于阅读时翻页
	if next != 0 {

		volumes, _ := QueryVolumes(chapter.BookID)
		log.Infof("Volumes: %#v", volumes)
		prevVolId, nextVolId := util.GetPrevNextId(volumes, chapter.VolumeID)

		log.Infof("prevVolId= %d, nextVolId= %d", prevVolId, nextVolId)

		// 下一篇
		nextc := &Chapter{}

		// 先检查同一卷内的
		has, err = x.Where("id>?", chapterid).And("book_id=?", chapter.BookID).And("volume_id=?", chapter.VolumeID).Asc("id").Get(nextc)
		if has && err == nil {
			data["next"] = map[string]any{"id": nextc.Id, "title": nextc.Title}
		} else if nextVolId != 0 {
			has, err = x.Where("book_id=?", chapter.BookID).And("volume_id=?", nextVolId).Asc("id").Get(nextc)
			if has && err == nil {
				data["next"] = map[string]any{"id": nextc.Id, "title": nextc.Title}
			}
		}

		prevc := &Chapter{}
		// 先检查同一卷内的
		has, err = x.Where("id<?", chapterid).And("book_id=?", chapter.BookID).And("volume_id=?", chapter.VolumeID).Desc("id").Get(prevc)
		if has && err == nil {
			data["prev"] = map[string]any{"id": prevc.Id, "title": prevc.Title}
		} else if prevVolId != 0 {
			has, err = x.Where("book_id=?", chapter.BookID).And("volume_id=?", prevVolId).Desc("id").Get(prevc)
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

	cols := "word_count"
	book := &Book{}

	if chapterid == 0 {
		_, err = sess.Insert(chapter)
		if err != nil {
			sess.Rollback()
			return 0, err
		}

		content.ChapterID = chapter.Id
		// content 表没有主键约束（真实库即为 chapterid BIGINT + txt TEXT），
		// xorm 在 Insert 后会把 ChapterID 回填成 content 表的隐式 rowid，
		// 所以先存一份真正的章 id 给索引用。
		newChapterID := content.ChapterID
		_, err = sess.Insert(content)
		if err != nil {
			sess.Rollback()
			return 0, err
		}

		// 同步正文索引；索引只是加速层，失败时告警但不阻断章节保存。
		if err = ftsUpsert(sess, newChapterID, content.Txt); err != nil {
			log.Warnf("写入章 %d 的正文索引失败：%v", newChapterID, err)
		}

		// 最新章节
		book.Latest = chapter.Title
		cols = "latest," + cols

		chapterid = chapter.Id
	} else {
		// 更新正文前先取旧内容：contentless 索引删除时必须提供原始 gram 文本。
		old := &Content{ChapterID: chapterid}
		if has, gerr := sess.Get(old); gerr != nil {
			log.Warnf("读取章 %d 的旧正文失败（索引可能残留）：%v", chapterid, gerr)
		} else if has {
			if err = ftsRemove(sess, chapterid, old.Txt); err != nil {
				log.Warnf("清理章 %d 的旧正文索引失败：%v", chapterid, err)
			}
		}

		if _, err = sess.ID(chapterid).Update(chapter); err != nil {
			sess.Rollback()
			return 0, err
		}

		_, err = sess.Update(content, &Content{ChapterID: chapterid})
		if err != nil {
			sess.Rollback()
			return 0, err
		}

		if err = ftsUpsert(sess, chapterid, content.Txt); err != nil {
			log.Warnf("更新章 %d 的正文索引失败：%v", chapterid, err)
		}
	}

	// 更新Book字数
	total, err := sess.Where("book_id=?", chapter.BookID).SumInt(new(Chapter), "word_count")
	if err != nil {
		sess.Rollback()
		return 0, err
	}
	book.WordCount = total

	if _, err = sess.ID(chapter.BookID).Cols(cols).Update(book); err != nil {
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
