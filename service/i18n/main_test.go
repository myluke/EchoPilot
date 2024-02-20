package i18n

import (
	"fmt"
	"testing"
)

func TestPrintf(t *testing.T) {
	ctx := Make("en")
	Printf(ctx, `test`)
	fmt.Println()
}
