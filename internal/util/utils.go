package util

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/unknwon/com"
)

type HasID interface {
	GetID() int64
}

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
	return fmt.Sprintf("%x", md5.Sum([]byte(pwd+salt)))
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

func SearchResult(cha, words string, leftLen int) string {
	length := len([]rune(words))

	content := ""

	if leftLen < 10 {
		leftLen = 10
	}

	for {
		idx, _ := StringIndexRune(cha, words)
		start := 0
		if idx > leftLen {
			start = idx - leftLen
		}

		if idx != -1 {
			if content == "" {
				content = Substring(cha, start, length+leftLen*2)
			} else {
				content = content + "..." + Substring(cha, start, length+leftLen*2)
			}

			cha = Substring(cha, idx+length, len([]rune(cha))-idx)
		} else {
			break
		}
	}

	return content + "..."
}

func StringIndexRune(s, substr string) (int, error) {
	byteIndex := strings.Index(s, substr)
	if byteIndex < 0 {
		return byteIndex, nil
	}
	reader := strings.NewReader(s)
	count := 0
	for byteIndex > 0 {
		_, bytes, err := reader.ReadRune()
		if err != nil {
			return 0, err
		}
		byteIndex = byteIndex - bytes
		count += 1
	}
	return count, nil
}

func Substring(input string, start int, length int) string {
	asRunes := []rune(input)

	if start >= len(asRunes) {
		return ""
	}

	if start+length > len(asRunes) {
		length = len(asRunes) - start
	}

	return string(asRunes[start : start+length])
}

// GetPrevNextId 泛型函数，查找目标 ID 的前一个和后一个 ID。
// 如果没有前一个或后一个，则返回 0。
func GetPrevNextId[T HasID](list []T, targetID int64) (prev, next int64) {
	for i, item := range list {
		if item.GetID() == targetID {
			if i > 0 {
				prev = list[i-1].GetID()
			}
			if i < len(list)-1 {
				next = list[i+1].GetID()
			}
			break
		}
	}
	return
}

// ParseRate 把类似 "1, 2,3 ,4" 这样的字符串，
// 转成 []int{1,2,3,4}。会自动去除前后空格、丢弃空项和非数字项。
func ParseRates(rate string) []int {
	var result []int

	// 直接分割，即使 rate 为空，也会得到 [""]，下面会被跳过
	parts := strings.Split(rate, ",")
	for _, part := range parts {
		// 去除前后空格
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// 尝试转换为整数，非纯数字（含字母、符号等）会出错并被丢弃
		if n, err := strconv.Atoi(part); err == nil {
			result = append(result, n)
		}
	}

	return result
}
