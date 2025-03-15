package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/CalebQ42/squashfs"
)

func userName(uid int, numeric bool) string {
	us := strconv.Itoa(uid)
	if numeric {
		return us
	}
	if u, err := user.LookupId(us); err == nil {
		return u.Username
	}
	return us
}

func groupName(gid int, numeric bool) string {
	gs := strconv.Itoa(gid)
	if numeric {
		return gs
	}
	if g, err := user.LookupGroupId(gs); err == nil {
		return g.Name
	}
	return gs
}

func printEntry(root, path string, d fs.DirEntry, numeric bool) {
	fi, _ := d.Info()
	sfi := fi.(squashfs.FileInfo)
	owner := fmt.Sprintf("%s/%s",
		userName(sfi.Uid(), numeric),
		groupName(sfi.Gid(), numeric))
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
	numeric := flag.Bool("lln", false, "List with attributes and numeric ids")
	offset := flag.Int64("o", 0, "Offset")
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
	r, err := squashfs.NewReaderAtOffset(f, *offset)
	if err != nil {
		panic(err)
	}
	if *list || *long || *numeric {
		root := flag.Arg(1)
		fs.WalkDir(r, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				panic(err)
			}
			if *long || *numeric {
				printEntry(root, path, d, *numeric)
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
