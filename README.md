# squashfs (WIP)

[![PkgGoDev](https://pkg.go.dev/badge/github.com/CalebQ42/squashfs)](https://pkg.go.dev/github.com/CalebQ42/squashfs) [![Go Report Card](https://goreportcard.com/badge/github.com/CalebQ42/squashfs)](https://goreportcard.com/report/github.com/CalebQ42/squashfs)

A PURE Go library to read and write squashfs.

Currently has support for reading squashfs files and extracting files and folders.

Special thanks to <https://dr-emann.github.io/squashfs/> for some VERY important information in an easy to understand format.
Thanks also to [distri's squashfs library](https://github.com/distr1/distri/tree/master/internal/squashfs) as I referenced it to figure some things out (and double check others).

## [TODO](https://github.com/CalebQ42/squashfs/projects/1?fullscreen=true)

## Limitations

This library is pure Go (including external libraries) which can cause some issues, which are listed below. Right now this library is also not feature complete, so check out the TODO list above for what I'm still planning on adding.

* No LZO compression. This is purely due to a lack of a good LZO pure golang library. If one is made, it would be a simple job to add it in.
* GZIP Compression
  * Strategies might or might not work.
  * Custom window sizes might or might not work.
* XZ Compression
  * LZMA executable filters are NOT supported.
* No Xattr parsing. This is simply because I haven't done any research on it and how to apply these in a pure go way.

## Performane

This library, decompressing the Firefox AppImage and using go tests, takes about twice as long as `unsquashfs` on my quad core laptop. (~1 second with the library and about half a second with `unsquashfs`)
