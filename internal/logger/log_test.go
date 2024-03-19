package logger

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gopkg.in/ini.v1"
)

const LIM = 100000

var file *ini.File

func initTestEnv() (err error) {
	if file == nil {
		file, err = ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, "/Users/zgia/program/projects/book-go/custom/conf/app.ini")
		if err != nil {
			return err
		}
		file.NameMapper = ini.SnackCase

		InitLogger(file, "", time.RFC3339Nano)
	}

	return nil
}

func TestDebugf(t *testing.T) {
	initTestEnv()

	for i := 0; i < LIM; i++ {
		Debugf("Welcome %s", "zgia")
	}
}

func TestSprintf(t *testing.T) {
	initTestEnv()

	for i := 0; i < LIM; i++ {
		fmt.Printf("Welcome %s", "zgia")
	}
}

func TestStrLength(t *testing.T) {
	str := "我们一起去abc"
	assert.Len(t, str, len(str), fmt.Sprintf("Length of str '%s' is %d", str, 18))
	assert.Equal(t, 8, len([]rune(str)))
}
