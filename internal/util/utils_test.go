package util

import (
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func init() {
	var cstZone = time.FixedZone("UTC", 8*3600) // 东八
	time.Local = cstZone
}

// go test -benchmem -run=^$ -bench=^BenchmarkRandStr$ -benchtime=200x -benchtime=2s -count=3 -cpuprofile=log/cpu.pprof zgia.net/book/internal/conf
func BenchmarkRandStr(b *testing.B) {
	for n := 0; n < b.N; n++ {
		RandStr(8, "ll")
	}
}

func TestXxx(t *testing.T) {

	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
	time.Local = time.FixedZone("UTC", 2*3600)
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))

	t.Run("", func(t *testing.T) {
		assert.Equal(t, 1, 1)
	})
}

func Test_Map(t *testing.T) {
	al := map[string]string{
		"l":   letters,
		"ln":  letterNumbers,
		"lln": liteLetterNumbers,
		"cl":  chaoticLetters,
	}
	alpha := al["x"]

	//fmt.Printf("%T, %T", alpha, nil)

	t.Run("", func(t *testing.T) {
		assert.Equal(t, "", alpha)
	})

}

func Test_RandStr(t *testing.T) {
	str := RandStr(6, "l")
	fmt.Println(str)
	t.Run("", func(t *testing.T) {
		assert.Equal(t, 6, len(str))
	})
}

func Test_StringTitle(t *testing.T) {

	st := strings.Title("hello world")
	real := "Hello World"
	t.Run("", func(t *testing.T) {
		assert.Equal(t, real, st)
	})

	st1 := cases.Title(language.Und).String("hello world")
	real1 := "Hello World"
	t.Run("", func(t *testing.T) {
		assert.Equal(t, real1, st1)
	})

}

func Test_String(t *testing.T) {
	const sample = "\xbd\xb2\x3d\xbc\x20\xe2\x8c\x98"
	real := "\xbd\xb2=\xbc ⌘"
	fmt.Printf("% x\t\t%q\t\t%+q\n", sample, sample, sample)

	t.Run("", func(t *testing.T) {
		assert.Equal(t, real, sample)
	})

	//printPlaceOfInterest()
	// china

}

func china() {
	const chinese = "中国人"
	for index, runeValue := range chinese {
		fmt.Printf("%#U, character %c, value %d, starts at byte position %d\n", runeValue, runeValue, runeValue, index)
	}

	for i, w := 0, 0; i < len(chinese); i += w {
		runeValue, width := utf8.DecodeRuneInString(chinese[i:])
		fmt.Printf("%U, character %c, starts at byte position %d\n", runeValue, runeValue, i)
		w = width
	}

	jp := []rune(chinese)
	l := utf8.RuneCountInString(chinese)
	fmt.Printf("length: %d - %d, %v\n", len(jp), l, jp)
	for _, r := range jp {
		fmt.Printf("%s(%v) ", string(r), r)
	}
	fmt.Println()

	const nihao = "nihao北京"
	l = utf8.RuneCountInString(nihao)
	fmt.Printf("Length of '%s' is %d, bytes length is %d \n", nihao, l, len(nihao))
}

func printPlaceOfInterest() {
	const placeOfInterest = `⌘`

	fmt.Printf("plain string: ")
	fmt.Printf("%s", placeOfInterest)
	fmt.Printf("\n")

	fmt.Printf("quoted string: ")
	fmt.Printf("%+q", placeOfInterest)
	fmt.Printf("\n")

	fmt.Printf("hex bytes: ")
	for i := 0; i < len(placeOfInterest); i++ {
		fmt.Printf("%x ", placeOfInterest[i])
	}
	fmt.Printf("\n")
}
