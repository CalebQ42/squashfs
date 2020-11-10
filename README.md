# GoSquashfs
My playground to mess around with Squashfs in Go. Might turn into an actual library someday. Mainly for AppImage

Right Now it's mostly based on [distri's squashfs library](https://github.com/distr1/distri/tree/master/internal/squashfs)

Special thanks to https://dr-emann.github.io/squashfs/ for some VERY important information in an easy to understand format

# Ideas
* Link directly to squashfs-tool using cgo
  * cgo can be a butt
  * Would require the least amount of code for me to do
  * Doing things in pure Go is definately highly superior
* Improve [distri's](https://github.com/distr1/distri) or [diskfs](https://github.com/diskfs/go-diskfs) squashfs library to meet my needs
  * Both don't work at all ATM.
  * distri's doesn't seem to support any sort of compression, which causes issues all over the place.
  * diskfs seems closer, but the squashfs code is incomplete
* Create an original squashfs library
  * MUCH more research needed
  * Could look at [squashfs-tools'](https://github.com/plougher/squashfs-tools) code to help out
  * Don't have to deal with other people's code :P
  * Honestly a bit unnecessary based on how close distri & diskfs's code is AND how little I know about squashfs.
