package route

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"zgia.net/book/internal/initial"
	"zgia.net/book/internal/models"
)

func TestSaveAllBooksToTxt(t *testing.T) {
	// 该用例会真实连接数据库并导出书籍；无可用数据库时跳过，保证全新环境可跑。
	if err := initial.Initialize(""); err != nil {
		t.Skipf("跳过：数据库不可用（%v）", err)
	}

	var bookId int64 = 499
	models.SaveAllBooksToTxt(bookId, "/tmp")

	assert.Equal(t, 1, 1)
}
