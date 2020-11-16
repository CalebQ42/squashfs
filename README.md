# GoSquashfs
My playground to mess around with Squashfs in Go. Might turn into an actual library someday. Mainly for AppImage

Right Now it's mostly based on [distri's squashfs library](https://github.com/distr1/distri/tree/master/internal/squashfs)

Special thanks to https://dr-emann.github.io/squashfs/ for some VERY important information in an easy to understand format

I am focusing purely on unsquashing before squashing. 

# Working

* Reading the header
* Reading data (slightly important :P)
* Reading inodes
* Reading directories
* Basic gzip compression (Shouldn't be too hard to implement other, but for right now, this works)

# Not Working (Yet). Roughly in order.

* Understanding the directory table. It's a bit weird TBH.
* Reading the UID, GUID, Xatt, Compression Options, Export, and Fragment tables.
* Implement other compression types
* Squashing

# Where I'm at

* Re-redid a bunch to try to make sure I wasn't durping. After that didn't work, I tried to figure out why things wheren't working, then realized HOW I was durping.