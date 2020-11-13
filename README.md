# GoSquashfs
My playground to mess around with Squashfs in Go. Might turn into an actual library someday. Mainly for AppImage

Right Now it's mostly based on [distri's squashfs library](https://github.com/distr1/distri/tree/master/internal/squashfs)

Special thanks to https://dr-emann.github.io/squashfs/ for some VERY important information in an easy to understand format

I am focusing purely on unsquashing before squashing. 

# Working

* Reading the header

# Not Working (Yet). Roughly in order.

* Actually reading the compressed data
* Reading Inodes
* Reading the Directory structure
* Implement other compression types
* Squashing

# Where I'm at

* Redid a bunch. Implemented a custom reader that can read across blocks.
    * As of yet, doesn't seem to be reading things quite right (seems to be issue with encryption reading)