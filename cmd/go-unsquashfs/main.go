package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/CalebQ42/squashfs"
	squashfslow "github.com/CalebQ42/squashfs/low"
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

var hardLinks = make(map[uint32]string)

func printFile(rdr *squashfs.Reader, path string, f *squashfs.File) {
	path = filepath.Join(path, f.Low.Name)
	fi, _ := f.Stat()
	sfi := fi.(squashfs.FileInfo)
	owner := fmt.Sprintf("%s/%s",
		userName(sfi.Uid(), *numeric),
		groupName(sfi.Gid(), *numeric))
	link, isHardLink := hardLinks[f.Low.Inode.Num]
	var size int64
	if isHardLink {
		size = 0
	} else {
		size = fi.Size()
		hardLinks[f.Low.Inode.Num] = path
	}
	if sfi.IsSymlink() {
		link = " -> " + sfi.SymlinkPath()
	} else if isHardLink {
		link = " link to " + link
	}
	fmt.Printf("%s %s %*d %s %s%s\n",
		strings.ToLower(fi.Mode().String()),
		owner, 26-len(owner), size,
		fi.ModTime().Format("2006-01-02 15:04"),
		path, link)
	if f.IsDir() {
		fs, _ := f.FS()
		printDir(rdr, path, fs)
	}
}

func printDir(rdr *squashfs.Reader, path string, f squashfs.FS) {
	var base squashfslow.FileBase
	var fil squashfs.File
	var err error
	for _, e := range f.LowDir.Entries {
		base, err = rdr.Low.BaseFromEntry(e)
		if err != nil {
			panic(err)
		}
		fil = rdr.FileFromBase(base, f)
		printFile(rdr, path, &fil)
	}
}

var (
	verbose       *bool
	list          *bool
	long          *bool
	numeric       *bool
	offset        *int64
	ignore        *bool
	file          *string
	showHardLinks *bool
)

func main() {
	verbose = flag.Bool("v", false, "Verbose")
	list = flag.Bool("l", false, "List")
	long = flag.Bool("ll", false, "List with attributes")
	numeric = flag.Bool("lln", false, "List with attributes and numeric ids")
	showHardLinks = flag.Bool("show-hard-links", false, "When used with ll or lln, shows hard links")
	offset = flag.Int64("o", 0, "Offset")
	ignore = flag.Bool("ip", false, "Ignore Permissions and extract all files/folders with 0755")
	file = flag.String("e", "", "File or folder to extract")
	flag.Parse()
	if (*list || *long || *numeric) && flag.NArg() < 1 {
		fmt.Println("Please provide a file name")
		os.Exit(0)
	} else if (!*list && !*long && !*numeric) && flag.NArg() < 2 {
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
	extractFil := r.File()
	if *file != "" {
		extractFil, err = r.OpenFile(*file)
		if err != nil {
			panic(err)
		}
	}
	if *list || *long || *numeric {
		printFile(&r, "", extractFil)
		return
	}
	op := squashfs.DefaultOptions()
	op.Verbose = *verbose
	op.IgnorePerm = *ignore
	n := time.Now()
	err = extractFil.ExtractWithOptions(flag.Arg(1), op)
	if err != nil {
		panic(err)
	}
	fmt.Println("Took:", time.Since(n))
}
