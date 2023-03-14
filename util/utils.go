package util

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"github.com/unknwon/com"
)

// EnsureAbs prepends the WorkDir to the given path if it is not an absolute path.
func EnsureAbs(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(WorkDir(), path)
}

type BookError struct {
	When time.Time
	What string
}

func (e *BookError) Error() string {
	return fmt.Sprintf("at %v, %s", e.When, e.What)
}

func Raise(msg string) error {
	return &BookError{time.Now(), msg}
}

// https://stackoverflow.com/a/31832326
// How to generate a random string of a fixed length in Go?
const (
	letters           = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterNumbers     = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	liteLetterNumbers = "0123456789abcdefghijklmnopqrstuvwxyz"
	chaoticLetters    = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ,./!@#$%^&*-_"
)

const (
	// 6 bits to represent a letter index
	letterIdBits = 6
	// All 1-bits as many as letterIdBits
	letterIdMask = 1<<letterIdBits - 1
	letterIdMax  = 63 / letterIdBits
)

// t: l=>letters, ln=>letterNumbers, lln=>liteLetterNumbers, cl=>chaoticLetters
func RandStr(n int, t string) string {

	var rndSrc = rand.NewSource(time.Now().UnixNano())

	alpha := letters
	switch t {
	case "ln":
		alpha = letterNumbers
	case "lln":
		alpha = liteLetterNumbers
	case "cl":
		alpha = chaoticLetters
	}

	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdMax letters!
	for i, cache, remain := n-1, rndSrc.Int63(), letterIdMax; i >= 0; {
		if remain == 0 {
			cache, remain = rndSrc.Int63(), letterIdMax
		}
		if idx := int(cache & letterIdMask); idx < len(alpha) {
			b[i] = alpha[idx]
			i--
		}
		cache >>= letterIdBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

// 使用salt对密码进行2次的md5
func EncodePassword(pwd string, salt string) string {
	p1 := fmt.Sprintf("%x", md5.Sum([]byte(pwd)))
	return fmt.Sprintf("%x", md5.Sum([]byte(p1+salt)))
}

// 页数不能小于 1
func PageNum(p string) int {
	return int(Max(ParamInt64(p), 1))
}

// ParamInt returns param result in int type.
func ParamInt(param string) int {
	return com.StrTo(param).MustInt()
}

func ParamInt64(param string) int64 {
	return com.StrTo(param).MustInt64()
}

func Min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int64) int64 {
	if x < y {
		return y
	}
	return x
}

func Eol(num int) string {

	e := "\n"
	if runtime.GOOS == "windows" {
		e = "\r\n"
	}

	if num < 2 {
		return e
	}

	return strings.Repeat(e, num)
}
