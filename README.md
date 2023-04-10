# squashfs

[![PkgGoDev](https://pkg.go.dev/badge/github.com/CalebQ42/squashfs)](https://pkg.go.dev/github.com/CalebQ42/squashfs) [![Go Report Card](https://goreportcard.com/badge/github.com/CalebQ42/squashfs)](https://goreportcard.com/report/github.com/CalebQ42/squashfs)

A PURE Go library to read squashfs. There is currently no plans to add archive creation support as it will almost always be better to just call `mksquashfs`. I could see some possible use cases, but probably won't spend time on it unless it's requested (open a discussion fi you want this feature).

Currently has support for reading squashfs files and extracting files and folders.

Special thanks to <https://dr-emann.github.io/squashfs/> for some VERY important information in an easy to understand format.
Thanks also to [distri's squashfs library](https://github.com/distr1/distri/tree/master/internal/squashfs) as I referenced it to figure some things out (and double check others).

## [TODO](https://github.com/CalebQ42/squashfs/projects/1?fullscreen=true)

## Limitations

* No Xattr parsing. This is simply because I haven't done any research on it and how to apply these in a pure go way.

## Performance

Testing on a zstd compressed file, my library is anywhere from 5x ~ 7x slower then `unsquashfs`
