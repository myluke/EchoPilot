package helper

import (
	"fmt"
	"testing"
)

func TestGender(t *testing.T) {
	a := "Sarah"
	// b := DetermineGender(a)
	ccc, eee := DetectGenderFromDict(a)
	// fmt.Println(b)
	fmt.Printf("%v, %s\n", ccc, eee)
}
