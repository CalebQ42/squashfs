# squashfs (WIP)

[![PkgGoDev](https://pkg.go.dev/badge/github.com/CalebQ42/squashfs)](https://pkg.go.dev/github.com/CalebQ42/squashfs) [![Go Report Card](https://goreportcard.com/badge/github.com/CalebQ42/squashfs)](https://goreportcard.com/report/github.com/CalebQ42/squashfs)

A PURE Go library to read and write squashfs.

Currently has support for reading squashfs files and extracting files and folders. Supports all compression types except LZO, but additional compression options are hit or miss.

The only major thing missing from squashfs reading is Xattr parsing.

Special thanks to https://dr-emann.github.io/squashfs/ for some VERY important information in an easy to understand format.
Thanks also to [distri's squashfs library](https://github.com/distr1/distri/tree/master/internal/squashfs) as I referenced it to figure some things out (and double check others).

# [TODO](https://github.com/CalebQ42/squashfs/projects/1?fullscreen=true)