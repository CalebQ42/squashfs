package squashfs

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/CalebQ42/squashfs/squashfs/inode"
)

const (
	squashfsURL  = "https://darkstorm.tech/files/LinuxPATest.sfs"
	squashfsName = "LinuxPATest.sfs"

	// filePath = "PortableApps/Notepad++Portable/App/DefaultData/Config/contextMenu.xml"
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
	err = checkDir(rdr, rdr.root)
	t.Fatal(err)
}

func checkDir(rdr *Reader, d *Directory) error {
	for _, e := range d.Entries {
		if e.InodeType == inode.Dir {
			b, err := d.Open(rdr, e.Name)
			if err != nil {
				return err
			}
			d, err := b.ToDir(rdr)
			if err != nil {
				return err
			}
			err = checkDir(rdr, d)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
