package squashfslow

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const (
	squashfsURL  = "https://darkstorm.tech/files/LinuxPATest.sfs"
	squashfsName = "airootfs.sfs"
)

func preTest(dir string) (fil *os.File, err error) {
	fil, err = os.Open(filepath.Join(dir, squashfsName))
	if err != nil {
		_, err = os.Open(dir)
		if os.IsNotExist(err) {
			err = os.Mkdir(dir, 0755)
		}
		if err != nil {
			return
		}
		os.Remove(filepath.Join(dir, squashfsName))
		fil, err = os.Create(filepath.Join(dir, squashfsName))
		if err != nil {
			return
		}
		var resp *http.Response
		resp, err = http.DefaultClient.Get(squashfsURL)
		if err != nil {
			return
		}
		_, err = io.Copy(fil, resp.Body)
		if err != nil {
			return
		}
	}
	_, err = exec.LookPath("unsquashfs")
	if err != nil {
		return
	}
	_, err = exec.LookPath("mksquashfs")
	return
}

func TestMisc(t *testing.T) {
	tmpDir := "../testing"
	fil, err := preTest(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer fil.Close()
	rdr, err := NewReader(fil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rdr.Superblock.FragCount)
	t.Fatal(rdr.fragEntry(1233))
}

func TestReader(t *testing.T) {
	tmpDir := "../testing"
	fil, err := preTest(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer fil.Close()
	rdr, err := NewReader(fil)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(tmpDir, "extractTest")
	os.RemoveAll(path)
	os.MkdirAll(path, 0777)
	err = extractToDir(rdr, rdr.Root.FileBase, path)
	if err != nil {
		t.Fatal(err)
	}
}

var singleFile = "PortableApps/CPU-X/CPU-X-v4.2.0-x86_64.AppImage"

func TestSingleFile(t *testing.T) {
	tmpDir := "../testing"
	fil, err := preTest(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer fil.Close()
	rdr, err := NewReader(fil)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(tmpDir, "extractTest")
	os.RemoveAll(path)
	os.MkdirAll(path, 0777)
	b, err := rdr.Root.Open(rdr, singleFile)
	if err != nil {
		t.Fatal(err)
	}
	err = extractToDir(rdr, b, path)
	if err != nil {
		t.Fatal(err)
	}
}

func extractToDir(rdr Reader, b FileBase, folder string) error {
	path := filepath.Join(folder, b.Name)
	if b.IsDir() {
		d, err := b.ToDir(rdr)
		if err != nil {
			return err
		}
		err = os.MkdirAll(path, 0777)
		if err != nil {
			return err
		}
		var nestBast FileBase
		for _, e := range d.Entries {
			nestBast, err = rdr.BaseFromEntry(e)
			if err != nil {
				return err
			}
			err = extractToDir(rdr, nestBast, path)
			if err != nil {
				return err
			}
		}
	} else if b.IsRegular() {
		_, full, err := b.GetRegFileReaders(rdr)
		if err != nil {
			return err
		}
		fil, err := os.Create(path)
		if err != nil {
			return err
		}
		_, err = full.WriteTo(fil)
		if err != nil {
			return err
		}
		fmt.Println("Successfully extracted file:", b.Name)
	}
	return nil
}
