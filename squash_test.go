package squashfs

import (
	"io"
	"net/http"
	"os"
	"testing"

	goappimage "github.com/CalebQ42/GoAppImage"
)

const (
	downloadURL  = "https://github.com/zilti/code-oss.AppImage/releases/download/continuous/Code_OSS-x86_64.AppImage"
	appImageName = "Code_OSS.AppImage"
)

func TestMain(t *testing.T) {
	t.Parallel()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	aiFil, err := os.Open(wd + "/testing/" + appImageName)
	if os.IsNotExist(err) {
		downloadTestAppImage(t, wd+"/testing")
		aiFil, err = os.Open(wd + "/testing/" + appImageName)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		t.Fatal(err)
	}
	defer aiFil.Close()
	stat, _ := aiFil.Stat()
	ai := goappimage.NewAppImage(wd + "/testing/" + appImageName)
	rdr, err := NewSquashfsReader(io.NewSectionReader(aiFil, ai.Offset, stat.Size()-ai.Offset))
	if err != nil {
		t.Fatal(err)
	}
	extractionFil := "code-oss.desktop"
	os.Remove(wd + "/testing/" + extractionFil)
	desk, err := os.Create(wd + "/testing/" + extractionFil)
	if err != nil {
		t.Fatal(err)
	}
	ext := rdr.FindFile(func(fil *File) bool {
		return fil.Name == extractionFil
	})
	if ext == nil {
		t.Fatal("Cannot find file")
	}
	_, err = io.Copy(desk, ext)
	if err != nil {
		t.Fatal(err)
	}
	t.Fatal("No problems here!")
}

func downloadTestAppImage(t *testing.T, dir string) {
	//seems to time out on slow connections. Might fix that at some point... or not
	os.Mkdir(dir, 0777)
	appImage, err := os.Create(dir + "/" + appImageName)
	if err != nil {
		t.Fatal(err)
	}
	defer appImage.Close()
	check := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	resp, err := check.Get(downloadURL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	_, err = io.Copy(appImage, resp.Body)
	if err != nil {
		t.Fatal(err)
	}
}
