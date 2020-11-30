package squashfs

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	goappimage "github.com/CalebQ42/GoAppImage"
	"github.com/CalebQ42/squashfs/internal/inode"
)

const (
	downloadURL  = "https://github.com/Swordfish90/cool-retro-term/releases/download/1.1.1/Cool-Retro-Term-1.1.1-x86_64.AppImage"
	appImageName = "Cool-Retro-Term.AppImage"
	squashfsName = "airootfs.sfs" //testing with a ArchLinux root fs from the live img
)

func TestSquashfs(t *testing.T) {
	t.Parallel()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	squashFil, err := os.Open(wd + "/testing/" + squashfsName)
	if err != nil {
		t.Fatal(err)
	}
	rdr, err := NewSquashfsReader(squashFil)
	if err != nil {
		t.Fatal(err)
	}
	fils, err := rdr.GetAllFiles()
	if err != nil {
		t.Fatal(err)
	}
	for _, fil := range fils {
		if fil.filType != inode.BasicFileType && fil.filType != inode.BasicDirectoryType && fil.filType != inode.BasicSymlinkType {
			fmt.Println("Found non-standard")
			fmt.Println(fil.Path())
			fmt.Println("type:", fil.filType)
		} else if fil.IsSymlink() {
			fmt.Println("Symlink!")
			fmt.Println(fil.Path())
			fmt.Println("symlink path:", fil.SymlinkPath())
		}
	}
}

func TestAppImage(t *testing.T) {
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
	} else if err != nil {
		t.Fatal(err)
	}
	defer aiFil.Close()
	stat, _ := aiFil.Stat()
	ai := goappimage.NewAppImage(wd + "/testing/" + appImageName)
	rdr, err := NewSquashfsReader(io.NewSectionReader(aiFil, ai.Offset, stat.Size()-ai.Offset))
	if err != nil {
		t.Fatal(err)
	}
	fil := rdr.GetFileAtPath("/usr/bin/cool-retro-term")
	if fil != nil {
		fmt.Println("Worked!", fil.Path())
	} else {
		t.Fatal("NOOOOOO!")
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

func TestCreateSquashFromAppImage(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	err = os.Mkdir(wd+"/testing", 0777)
	if err != nil && !os.IsExist(err) {
		t.Fatal(err)
	}
	_, err = os.Open(wd + "/testing/" + appImageName)
	if os.IsNotExist(err) {
		downloadTestAppImage(t, wd+"/testing")
		_, err = os.Open(wd + "/testing/" + appImageName)
		if err != nil {
			t.Fatal(err)
		}
	} else if err != nil {
		t.Fatal(err)
	}
	ai := goappimage.NewAppImage(wd + "/testing/" + appImageName)
	aiFil, err := os.Open(wd + "/testing/" + appImageName)
	if err != nil {
		t.Fatal(err)
	}
	defer aiFil.Close()
	aiFil.Seek(ai.Offset, 0)
	os.Remove(wd + "/testing/" + appImageName + ".squashfs")
	aiSquash, err := os.Create(wd + "/testing/" + appImageName + ".squashfs")
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.Copy(aiSquash, aiFil)
	if err != nil {
		t.Fatal(err)
	}
}
