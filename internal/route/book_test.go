package route

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"zgia.net/book/internal/initial"
	_ "zgia.net/book/testify"
)

func TestSaveAllBooksToTxt(t *testing.T) {
	err := initial.Initialize("")
	if err != nil {
		t.Fatal(err)
	}

	all, books := SaveAllBooksToTxt()

	assert.Equal(t, all, books)
}
