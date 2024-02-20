package helper

import (
	"fmt"
	"log"
	"regexp"
	"testing"
)

func TestFindSubstrByString(t *testing.T) {
	content := `func (re *Regexp) FindAllIndex(b []byte, n int) [][]int`
	resp := FindSubstr(content, "(", ")")
	result := "re *Regexp"
	if resp != result {
		t.Errorf("Error: %s!=%s", resp, result)
		t.Fatal()
	}
	fmt.Println(resp)
}

func TestFindSubstrByRegexp(t *testing.T) {
	content := `func (re *Regexp) FindAllIndex(b []byte, n int) [][]int`
	resp := FindSubstr(content, regexp.MustCompile(`c\s*\(`), regexp.MustCompile(`\)`))
	result := "re *Regexp"
	if resp != result {
		t.Errorf("Error: %s!=%s", resp, result)
		t.Fatal()
	}
	fmt.Println(resp)
}

func TestFindSubstrByMixing(t *testing.T) {
	content := `func (re *Regexp) FindAllIndex(b []byte, n int) [][]int`
	resp := FindSubstr(content, regexp.MustCompile(`c\s*\(`), ")")
	result := "re *Regexp"
	if resp != result {
		t.Errorf("Error: %s!=%s", resp, result)
		t.Fatal()
	}
	fmt.Println(resp)
}

func TestStrLen(t *testing.T) {
	var r int
	r = StrLen("hello word")
	log.Printf("hello word: %d\n", r)
	r = StrLen("A")
	log.Printf("A: %d\n", r)
	r = StrLen("1")
	log.Printf("1: %d\n", r)
	r = StrLen("中")
	log.Printf("中: %d\n", r)
	r = StrLen("中文")
	log.Printf("中文: %d\n", r)
	r = StrLen("😀")
	log.Printf("😀: %d\n", r)
	r = StrLen("1😀")
	log.Printf("1😀: %d\n", r)
	r = StrLen("中文😀")
	log.Printf("中文😀: %d\n", r)
	r = StrLen("中👮‍♂️")
	log.Printf("中👮‍♂️: %d\n", r)
}
