package models

import (
	"zgia.net/book/internal/db"
)

// 接口返回
type VolumeResult struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Cover   string `json:"cover"`
}

func ListVolumes(book *db.Book) (map[string]any, error) {

	volumes, err := db.QueryVolumes(book.Id)

	if err != nil {
		return nil, err
	}

	vr := make([]*VolumeResult, len(volumes))

	for i, v := range volumes {
		vr[i] = GetVolume(v)
	}

	dt := map[string]any{
		"items": vr,
		"book":  map[string]string{"title": book.Title},
	}

	return dt, nil
}

func GetVolume(vol *db.Volume) *VolumeResult {

	return &VolumeResult{
		Id:      vol.Id,
		Title:   vol.Title,
		Summary: vol.Summary,
		Cover:   vol.Cover,
	}
}
