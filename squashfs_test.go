package squashfs_test

//Actually proper tests go here.

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/CalebQ42/squashfs"
)

const (
	squashfsURL  = "https://darkstorm.tech/LinuxPATest.sfs"
	squashfsName = "LinuxPATest.sfs"
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

func BenchmarkRace(b *testing.B) {
	// tmpDir := b.TempDir()
	tmpDir := "testing"
	fil, err := preTest(tmpDir)
	if err != nil {
		b.Fatal(err)
	}
	libPath := filepath.Join(tmpDir, "ExtractLib")
	unsquashPath := filepath.Join(tmpDir, "ExtractSquashfs")
	os.RemoveAll(libPath)
	os.RemoveAll(unsquashPath)
	var libTime, unsquashTime time.Duration
	start := time.Now()
	rdr, err := squashfs.NewReader(fil)
	if err != nil {
		b.Fatal(err)
	}
	err = rdr.ExtractTo(libPath)
	if err != nil {
		b.Fatal(err)
	}
	libTime = time.Since(start)
	cmd := exec.Command("unsquashfs", "-d", unsquashPath, fil.Name())
	start = time.Now()
	err = cmd.Run()
	if err != nil {
		b.Fatal(err)
	}
	unsquashTime = time.Since(start)
	b.Log("Library took:", libTime.Round(time.Millisecond))
	b.Log("unsquashfs took:", unsquashTime.Round(time.Millisecond))
	b.Log("unsquashfs is", strconv.FormatFloat(float64(libTime.Milliseconds())/float64(unsquashTime.Milliseconds()), 'f', 2, 64), "times faster")
}

func TestExtractQuick(t *testing.T) {

	//First, setup everything and extract the archive using the library and unsquashfs

	// tmpDir := b.TempDir()
	tmpDir := "testing"
	fil, err := preTest(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	libPath := filepath.Join(tmpDir, "ExtractLib")
	unsquashPath := filepath.Join(tmpDir, "ExtractSquashfs")
	os.RemoveAll(libPath)
	os.RemoveAll(unsquashPath)
	rdr, err := squashfs.NewReader(fil)
	if err != nil {
		t.Fatal(err)
	}
	op := squashfs.DefaultOptions()
	op.Verbose = true
	err = rdr.ExtractWithOptions(libPath, op)
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
	err = fs.WalkDir(squashFils, ".", func(path string, d fs.DirEntry, _ error) error {
		libFil, e := os.Open(filepath.Join(libPath, path))
		if e != nil {
			return e
		}
		stat, _ := d.Info()
		libStat, _ := libFil.Stat()
		if stat.Size() != libStat.Size() {
			t.Log(path, "not the same size between library and unsquashfs")
			t.Log("File is", libStat.Size())
			t.Log("Should be", stat.Size())
			return errors.New("file not the correct size")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
