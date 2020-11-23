# GoSquashfs

A PURE Go library to read and write squashfs. Right now I'm focusing on unsquashing.
Currently IS NOT a functional library. Some things that are currently public IS going to become private. Not very well documented either.

Special thanks to https://dr-emann.github.io/squashfs/ for some VERY important information in an easy to understand format.
Thanks also to [distri's squashfs library](https://github.com/distr1/distri/tree/master/internal/squashfs) as I referenced it to figure some things out (and double check others).

# Working

* Reading the header
* Reading metadata blocks (whether encrypted or not)
* Reading inodes
* Reading directories
* Basic gzip compression (Shouldn't be too hard to implement other, but for right now, this works)
* Listing all files via a string slice

# Not Working (Yet). Roughly in order.

* Figure out fragments (I can't seem to make them work ATM)
* Extracting files
    * from inodes.
    * from path.
    * from file info.
* Give a list of files
    * In io.FileStat (?) form
* Reading the UID, GUID, Xatt, Compression Options, Export, and Fragment tables.
* Implement other compression types (Should be relatively easy)
* Squashing

# Where I'm at.

* I've given up on fragments for now and will work on reading files.
    * Once I have basic file reading working, I'll have a first pre-release available. No fragment support, but shouldn't be too hard... right?