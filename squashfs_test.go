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
	tmpDir := "testing"
	fil, err := preTest(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	rdr, err := squashfs.NewReader(fil)
	if err != nil {
		t.Fatal(err)
	}
	_ = rdr
	// Put testing here
	// t.Fatal("UM")
}

func BenchmarkRace(b *testing.B) {
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
	op := squashfs.FastOptions()
	op.IgnorePerm = true
	start := time.Now()
	rdr, err := squashfs.NewReader(fil)
	if err != nil {
		b.Fatal(err)
	}
	err = rdr.ExtractWithOptions(libPath, op)
	if err != nil {
		b.Fatal(err)
	}
	libTime = time.Since(start)
	cmd := exec.Command("unsquashfs", "-d", unsquashPath, fil.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	start = time.Now()
	err = cmd.Run()
	if err != nil {
		b.Log("Unsquashfs error:", err)
	}
	unsquashTime = time.Since(start)
	// b.Log("Library took:", libTime.Round(time.Millisecond))
	// b.Log("unsquashfs took:", unsquashTime.Round(time.Millisecond))
	b.Fatal("unsquashfs is", strconv.FormatFloat(float64(libTime.Milliseconds())/float64(unsquashTime.Milliseconds()), 'f', 2, 64), "times faster")
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
	os.RemoveAll(filepath.Join(tmpDir, "testLog.txt"))
	logFil, _ := os.Create(filepath.Join(tmpDir, "testLog.txt"))
	op := squashfs.DefaultOptions()
	op.Verbose = true
	op.IgnorePerm = true
	op.LogOutput = logFil
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
	err = fs.WalkDir(squashFils, ".", func(path string, _ fs.DirEntry, _ error) error {
		libFil, e := os.Open(filepath.Join(libPath, path))
		if e != nil {
			return e
		}
		sfsFile, e := os.Open(filepath.Join(unsquashPath, path))
		if e != nil {
			return e
		}
		sfsStat, _ := sfsFile.Stat()
		libStat, _ := libFil.Stat()
		if sfsStat.Size() != libStat.Size() {
			t.Log(libFil.Name(), "not the same size between library and unsquashfs")
			t.Log("File is", libStat.Size())
			t.Log("Should be", sfsStat.Size())
			return errors.New("file not the correct size")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

var filePath = "bin"

func TestSingleFile(t *testing.T) {
	tmpDir := "testing"
	fil, err := preTest(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	os.Remove(filepath.Join(tmpDir, filePath))
	rdr, err := squashfs.NewReader(fil)
	if err != nil {
		t.Fatal(err)
	}
	f, err := rdr.Open(filePath)
	if err != nil {
		t.Fatal(err)
	}
	err = f.(*squashfs.File).ExtractWithOptions("testing", &squashfs.ExtractionOptions{Verbose: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Fatal("HI")
}
