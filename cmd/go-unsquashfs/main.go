package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/CalebQ42/squashfs"
)

func main() {
	verbose := flag.Bool("v", false, "Verbose")
	list := flag.Bool("l", false, "List")
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
	if *list {
		root := flag.Arg(1)
		fs.WalkDir(r, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				panic(err)
			}
			fmt.Println(filepath.Join(root, path))
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
