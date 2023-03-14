package db

import (
	"fmt"
)

type Volume struct {
	Id        int64 `xorm:"pk autoincr"`
	Bookid    int64 `xorm:"notnull default 0"`
	Title     string
	Summary   string
	Createdat int64 `xorm:"created notnull default 0"`
	Updatedat int64 `xorm:"updated notnull default 0"`
	Deletedat int64 `xorm:"deleted notnull default 0"`
}

// 接口返回
type VolumeResult struct {
	Id      int64
	Title   string
	Summary string
}

func (b *Volume) String() string {
	return fmt.Sprintf("Volume: %v, bookid: %d, %v", b.Title, b.Bookid, b.Summary)
}

// ListVolumes returns number of volumes in given page.
func QueryVolumes(bookid int64) ([]*Volume, error) {
	var volumes []*Volume
	return volumes, x.Where("bookid=?", bookid).And("deletedat=?", 0).Asc("id").Find(&volumes)
}

// UpdateVolume updates/creates a volume
func UpdateVolume(volumeid int64, volume *Volume) (id int64, err error) {

	sess := x.NewSession()
	defer sess.Close()

	if err = sess.Begin(); err != nil {
		return 0, err
	}

	if volumeid == 0 {
		_, err = sess.Insert(volume)
		if err != nil {
			return 0, err
		}

		volumeid = volume.Id
	} else {
		if _, err = sess.ID(volumeid).Cols("title", "summary").Update(volume); err != nil {
			return 0, err
		}
	}

	if err = sess.Commit(); err != nil {
		return 0, err
	}

	return volumeid, nil
}

// DeleteVolume deletes a Volume
func DeleteVolume(volumeid int64) (int64, error) {
	return x.ID(volumeid).Delete(new(Volume))
}
