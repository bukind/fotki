Playing with golang as a simple project of a photo album.

All photos are (re)-organized into a hierarchy:

```
YYYY/
  all/YYYY-MM/IMG
  YYYY-MM-DD[-suffix]/IMG
```

At the top level, there are years.  Inside each year there are days
subfolders and also a special folder all/ which contains months.
Months folders are named in the form `YYYY-MM`.
The day folder begins with `YYYY-MM-DD` and may have an optional
suffix.

The date of a photo is found using:
  * from the EXIF data embedded into the photo;
  * by parsing the name of the photo looking for patterns like
    YYYYMMDD, etc.
  * if the photo name is a single integer number representing
    a timestamp in milliseconds since Epoch.

Then the photo is placed into both month and day folder using the
hardlink.  Thus the disk space is not wasted.

The tool can be run as:

```
$ fotki --root ROOTDIR --scan SCANDIR [-v] [-n]
```

where the `ROOTDIR` is the root of the album, and the `SCANDIR` is
the directory to start scanning.


Feature requests:
  * Remove original photos after they are moved into destination folders
  * DONE: Avoid descending to the canonical folders (months, days) while scanning
  * NOT NEEDED: Collect md5sum of images to eliminate duplicates
  * Identify garbage and optionally remove it.
  * DONE: Store FileInfo in dir while scanning to optimize disk access
