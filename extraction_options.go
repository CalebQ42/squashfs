package squashfs

import (
	"io"
	"io/fs"
)

type ExtractionOptions struct {
	LogOutput          io.Writer   //Where error log should write.
	DereferenceSymlink bool        //Replace symlinks with the target file.
	UnbreakSymlink     bool        //Try to make sure symlinks remain unbroken when extracted, without changing the symlink.
	Verbose            bool        //Prints extra info to log on an error.
	IgnorePerm         bool        //Ignore file's permissions and instead use Perm.
	Perm               fs.FileMode //Permission to use when IgnorePerm. Defaults to 0777.
}

func DefaultOptions() *ExtractionOptions {
	return &ExtractionOptions{
		Perm: 0777,
	}
}
