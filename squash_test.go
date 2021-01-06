package squashfs

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	goappimage "github.com/CalebQ42/GoAppImage"
)

const (
	downloadURL  = "https://github.com/srevinsaju/Firefox-Appimage/releases/download/firefox-v84.0.r20201221152838/firefox-84.0.r20201221152838-x86_64.AppImage"
	appImageName = "firefox-84.0.r20201221152838-x86_64.AppImage"
	squashfsName = "balenaEtcher-1.5.113-x64.AppImage.sfs" //testing with a ArchLinux root fs from the live img
)

func TestSquashfs(t *testing.T) {
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
	fmt.Println("stuff", rdr.super.CompressionType)
	fil := rdr.GetFileAtPath("*.desktop")
	if fil == nil {
		t.Fatal("Can't find desktop fil")
	}
	errs := fil.ExtractTo(wd + "/testing")
	if len(errs) > 0 {
		t.Fatal(errs)
	}
	errs = rdr.ExtractTo(wd + "/testing/" + squashfsName + ".d")
	if len(errs) > 0 {
		t.Fatal(errs)
	}
	t.Fatal("No Problems")
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
	os.RemoveAll(wd + "testing/firefox")
	errs := rdr.ExtractTo(wd + "/testing/firefox")
	if len(errs) > 0 {
		t.Fatal(errs)
	}
	// os.RemoveAll(wd + "/testing/" + appImageName + ".d")
	// root, _ := rdr.GetRootFolder()
	// errs := root.ExtractWithOptions(wd+"/testing/"+appImageName+".d", true, os.ModePerm, true)
	// t.Fatal(errs)
	t.Fatal("No problemo!")
}

func downloadTestAppImage(t *testing.T, dir string) {
	//seems to time out on slow connections. Might fix that at some point... or not. It's just a test...
	os.Mkdir(dir, os.ModePerm)
	appImage, err := os.Create(dir + "/" + appImageName)
	if err != nil {
		t.Fatal(err)
	}
	defer appImage.Close()
	check := http.Client{
		CheckRedirect: func(r *http.Request, _ []*http.Request) error {
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
