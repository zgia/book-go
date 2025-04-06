package db

import "fmt"

type Volume struct {
	Id        int64 `xorm:"pk autoincr"`
	Bookid    int64 `xorm:"notnull default 0"`
	Title     string
	Summary   string
	Cover     string
	Createdat int64 `xorm:"created notnull default 0"`
	Updatedat int64 `xorm:"updated notnull default 0"`
	Deletedat int64 `xorm:"deleted notnull default 0"`
}

func (v *Volume) GetID() int64 {
	return v.Id
}

func (b *Volume) String() string {
	return fmt.Sprintf("Volume: %v, bookid: %d, %v", b.Title, b.Bookid, b.Summary)
}

// ListVolumes returns number of volumes in given page.
func QueryVolumes(bookid int64) ([]*Volume, error) {
	var volumes []*Volume
	return volumes, x.Where("bookid=?", bookid).Asc("id").Find(&volumes)
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
	return x.Delete(Volume{Id: volumeid})
}
