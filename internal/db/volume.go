package db

import (
	"fmt"
	"time"
)

type Volume struct {
	Id        int64 `xorm:"pk autoincr"`
	BookID    int64 `xorm:"notnull default 0"`
	Title     string
	Summary   string
	Cover     string
	CreatedAt time.Time `xorm:"created notnull"`
	UpdatedAt time.Time `xorm:"updated notnull"`
	DeletedAt time.Time `xorm:"deleted null"`
}

func (v *Volume) GetID() int64 {
	return v.Id
}

func (b *Volume) String() string {
	return fmt.Sprintf("Volume: %v, book_id: %d, %v", b.Title, b.BookID, b.Summary)
}

// ListVolumes returns number of volumes in given page.
func QueryVolumes(bookid int64) ([]*Volume, error) {
	var volumes []*Volume
	return volumes, x.Where("book_id=?", bookid).Asc("id").Find(&volumes)
}

// ListVolumes returns number of volumes in given page.
func QueryVolume(volumeid int64) (*Volume, error) {
	volume := &Volume{Id: volumeid}
	has, err := x.Get(volume)

	if err != nil {
		return nil, err
	}

	if !has {
		return nil, nil
	}

	return volume, nil
}

// UpdateVolume updates/creates a volume
func UpdateVolume(volumeid int64, volume *Volume) (int64, error) {
	var err error

	sess := x.NewSession()
	defer sess.Close()

	if err = sess.Begin(); err != nil {
		return 0, err
	}

	if volumeid == 0 {
		_, err = sess.Insert(volume)
		if err != nil {
			sess.Rollback()
			return 0, err
		}

		volumeid = volume.Id
	} else {
		if _, err = sess.ID(volumeid).Cols("title", "summary").Update(volume); err != nil {
			sess.Rollback()
			return 0, err
		}
	}

	if err = sess.Commit(); err != nil {
		sess.Rollback()
		return 0, err
	}

	return volumeid, nil
}

// DeleteVolume deletes a Volume
func DeleteVolume(volumeid int64) (int64, error) {
	return x.Delete(Volume{Id: volumeid})
}
