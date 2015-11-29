# box
A rootless client that allows efficient data syncing with a remote server using rsync-like protocol

- [Awesome Data Structures](http://stackoverflow.com/questions/500607/what-are-the-lesser-known-but-useful-data-structures)

On Walking the FS:

- [in c](http://stackoverflow.com/questions/2312110/efficiently-traverse-directory-tree-with-opendir-readdir-and-closedir)

Bloom Filters:

- [in Venti](http://essay.utwente.nl/694/1/scriptie_Lukkien.pdf):
  Venti uses an in-memory bloom filter to avoid
having to go to index section disks to determine whether a score is present. This
is most important for writing new blocks to Venti: new blocks will not be found
in the index, it would be nice not to have to read the index, which takes a full
disk seek. A bloom filter can give false positives in the membership test, but not
false negatives. Therefore, when the bloom filter determines the score is absent,
it really is absent. When the bloom filter determines the score is present, it may
be absent after all (this will be realised after having read the index).
- [calculator](http://hur.st/bloomfilter?n=100000000000000&p=0.01)

On Rsync:

- [list of docs](https://rsync.samba.org/documentation.html)
- [how rsync works](https://rsync.samba.org/how-rsync-works.html)
- [with merkle tree](http://blog.kodekabuki.com/post/11135148692/rsync-internals)
- [rolling explained](http://stackoverflow.com/questions/12456523/how-to-generate-rolling-checksum-for-overlapping-chunks)

On Dropbox:

- [How Dropbox works](http://stackoverflow.com/questions/185533/how-does-the-dropbox-mac-client-work)
- [reversed engineered](http://archive.hack.lu/2012/Dropbox%20security.pdf)
- [dropbox's librsync](https://github.com/dropbox/librsync)
- [inside dropbox](http://cnds.eecs.jacobs-university.de/courses/nds-2013/prodescu-inside-dropbox.pdf)

Merkle Tree:

- [for syncing state between nodes](http://danieloshea.com/2011/12/07/merkle-tree.html)
- [Optimization and Hashes](http://crypto.stackexchange.com/questions/9198/efficient-incremental-updates-to-large-merkle-tree)
