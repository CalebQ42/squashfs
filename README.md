# GoSquashfs

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

# Not Working (Yet). Roughly in order.

* Reading the UID, GUID, Xatt, Compression Options, and Export tables.
* Extracting files
    * from inodes.
    * from file info.
* Give a list of files
    * In io.FileStat (?) form
* Reading the UID, GUID, Xatt, Compression Options, and Export tables.
* Implement other compression types (Should be relatively easy)
* Squashing
* Threading processes to speed them up

# Where I'm at.

* I FINALLY GOT FILE EXTRACTION WORKING!!