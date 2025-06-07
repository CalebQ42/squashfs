package squashfs

import (
	"io"
	"io/fs"
	"runtime"
	"sync"
)

type ExtractionOptions struct {
	dispatcher         chan struct{} // Limits the amount of work being done simultaneously.
	fullRdrPool        sync.Pool     // Pool for data.FullReader results.
	LogOutput          io.Writer     //Where the verbose log should write.
	DereferenceSymlink bool          //Replace symlinks with the target file.
	UnbreakSymlink     bool          //Try to make sure symlinks remain unbroken when extracted, without changing the symlink.
	Verbose            bool          //Prints extra info to log on an error.
	IgnorePerm         bool          //Ignore file's permissions and instead use Perm.
	Perm               fs.FileMode   //Permission to use when IgnorePerm. Defaults to 0777.
	ExtractionRoutines uint16        //The number of threads to use during extraction. Defaults to a number based on runtime.NumCPU().
	SimultaneousFiles  uint16        //Depreciated: Only use ExtractionRoutines
}

// The default extraction options. Uses half of your CPU cores.
func DefaultOptions() *ExtractionOptions {
	return &ExtractionOptions{
		Perm:               0777,
		ExtractionRoutines: uint16(runtime.NumCPU() / 2),
	}
}

// Faster extraction option. Uses all CPU cores.
func FastOptions() *ExtractionOptions {
	return &ExtractionOptions{
		Perm:               0777,
		ExtractionRoutines: uint16(runtime.NumCPU()),
	}
}
