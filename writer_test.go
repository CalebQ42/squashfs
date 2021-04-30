package squashfs

import (
	"os"
	"testing"
)

func TestWrite(t *testing.T) {
	os.Remove("testing/test.sfs")
	os.Mkdir("testing", os.ModePerm)
	test, err := os.Create("testing/test.sfs")
	if err != nil {
		t.Fatal(err)
	}
	_ = test
}
