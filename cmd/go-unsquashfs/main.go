package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/CalebQ42/squashfs"
)

func printEntry(root, path string, d fs.DirEntry) {
	fi, _ := d.Info()
	sfi := fi.(squashfs.FileInfo)
	owner := fmt.Sprintf("%d/%d",
		sfi.Uid(),
		sfi.Gid())
	link := ""
	if sfi.IsSymlink() {
		link = " -> " + sfi.SymlinkPath()
	}
	fmt.Printf("%s %s %*d %s %s%s\n",
		strings.ToLower(fi.Mode().String()),
		owner, 26-len(owner), fi.Size(),
		fi.ModTime().Format("2006-01-02 15:04"),
		filepath.Join(root, path), link)
}

func main() {
	verbose := flag.Bool("v", false, "Verbose")
	list := flag.Bool("l", false, "List")
	long := flag.Bool("ll", false, "List with attributes")
	ignore := flag.Bool("ip", false, "Ignore Permissions and extract all files/folders with 0755")
	flag.Parse()
	if len(flag.Args()) < 2 {
		fmt.Println("Please provide a file name and extraction path")
		os.Exit(0)
	}
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		panic(err)
	}
	r, err := squashfs.NewReader(f)
	if err != nil {
		panic(err)
	}
	if *list || *long {
		root := flag.Arg(1)
		fs.WalkDir(r, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				panic(err)
			}
			if *long {
				printEntry(root, path, d)
			} else {
				fmt.Println(filepath.Join(root, path))
			}
			return nil
		})
		return
	}
	op := squashfs.DefaultOptions()
	op.Verbose = *verbose
	op.IgnorePerm = *ignore
	n := time.Now()
	err = r.ExtractWithOptions(flag.Arg(1), op)
	if err != nil {
		panic(err)
	}
	fmt.Println("Took:", time.Since(n))
}
