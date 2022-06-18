package squashfs_test

import (
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/CalebQ42/squashfs"
)

const (
	squashfsURL  = "https://darkstorm.tech/LinuxPATest.sfs"
	squashfsName = "LinuxPATest.sfs"
)

func preTest(dir string) (fil *os.File, err error) {
	_, err = os.Open(filepath.Join(dir, squashfsName))
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

func TestExtractQuick(t *testing.T) {

	//First, setup everything and extract the archive using the library and unsquashfs

	tmpDir := t.TempDir()
	fil, err := preTest(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	libPath := filepath.Join(tmpDir, "ExtractLib")
	unsquashPath := filepath.Join(tmpDir, "ExtractSquashfs")
	os.Remove(libPath)
	os.Remove(unsquashPath)
	rdr, err := squashfs.NewReader(fil)
	if err != nil {
		t.Fatal(err)
	}
	err = rdr.ExtractTo(libPath)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("unsquashfs", "-d", unsquashPath, fil.Name())
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	//Then compare the sizes and existance between the two (using unsquashfs as a reference).
	//If the file doesn't exist, or the size is different, we exit.
	//TODO: Add long test that checks contents.

	squashFils := os.DirFS(unsquashPath)
	err = fs.WalkDir(squashFils, "", func(path string, d fs.DirEntry, err error) error {
		t.Log(path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Fatal("This is a test")
}
