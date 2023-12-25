package squashfs

import (
	"io"
	"io/fs"
	"os"

	"github.com/CalebQ42/squashfs/internal/routinemanager"
)

type ExtractionOptions struct {
	manager            *routinemanager.Manager
	LogOutput          io.Writer   //Where the verbose log should write. Defaults to os.Stdout.
	DereferenceSymlink bool        //Replace symlinks with the target file.
	UnbreakSymlink     bool        //Try to make sure symlinks remain unbroken when extracted, without changing the symlink.
	Verbose            bool        //Prints extra info to log on an error.
	IgnorePerm         bool        //Ignore file's permissions and instead use Perm.
	Perm               fs.FileMode //Permission to use when IgnorePerm. Defaults to 0777.
	SimultaneousFiles  uint16      //Number of files to process in parallel. Defaults to 10.
	ExtractionRoutines uint16      //Number of goroutines to use for each file's extraction. Only applies to regular files. Defaults to 10.
}

func DefaultOptions() *ExtractionOptions {
	return &ExtractionOptions{
		LogOutput:          os.Stdout,
		Perm:               0777,
		SimultaneousFiles:  10,
		ExtractionRoutines: 10,
	}
}
