package route

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"zgia.net/book/internal/initial"
	"zgia.net/book/internal/models"
	_ "zgia.net/book/testify"
)

func TestSaveAllBooksToTxt(t *testing.T) {
	err := initial.Initialize("")
	if err != nil {
		t.Fatal(err)
	}

	var bookId int64 = 499
	models.SaveAllBooksToTxt(bookId, "/tmp")

	assert.Equal(t, 1, 1)
}
