# GoSquashfs

[![PkgGoDev](https://pkg.go.dev/badge/github.com/CalebQ42/GoSquashfs)](https://pkg.go.dev/github.com/CalebQ42/GoSquashfs)

A PURE Go library to read and write squashfs.

Currently, you can read a squashfs and extract files (only files at the moment). Many things are public that shouldn't be, but you can use it by using NewSquashfsReader and subsequent ReadFile.

Special thanks to https://dr-emann.github.io/squashfs/ for some VERY important information in an easy to understand format.
Thanks also to [distri's squashfs library](https://github.com/distr1/distri/tree/master/internal/squashfs) as I referenced it to figure some things out (and double check others).

# Working

* Extracting files from string paths
* Reading the header
* Reading metadata blocks (whether encrypted or not)
* Reading inodes
* Reading directories
* Basic gzip compression (Shouldn't be too hard to implement other, but for right now, this works)
* Listing all files via a string slice

# Not Working (Yet). Not necessarily in order.

* Provide an easy interface to find and list files and their properties
    * Maybe squashfs.File
* Make device, socket, symlink, and all extended types of inode work properly. (I need to find an archive that uses it first.)
* Extracting files
    * from inodes.
    * from file info.
* Give a list of files
    * In io.FileStat (?) form
* Reading the UID, GUID, Xatt, Compression Options, and Export tables.
* Implement other compression types (Should be relatively easy)
* Squashing
* Threading processes to speed them up
* Reasonable tests

# TODO

* Go over all documentation again (especially for exported structs and functions) to make sure it's easy to understand.

# Where I'm at

* Working on the File interface that should make it easier to deal with squashfs files. I'm also trying to make them capable for when I get squashing working.