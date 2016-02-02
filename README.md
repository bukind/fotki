Playing with golang as a simple project of a photo album.

All photos are (re)-organized into a hierarchy:

YYYY/
  YYYY-MM/
  all/YYYY-MM-DD[-suffix]/

At the top level, there are years.  Inside each year there are months
in the form of YYYY-MM and also a special folder all/ which contains
days.  The day folder begins with YYYY-MM-DD and may have an optional
suffix.

The date of a photo is found using:
1. from the EXIF data embedded into the photo;
2. by parsing the name of the photo looking for patterns like
   YYYYMMDD, etc.

Then the photo is placed into both month and day folder using the
hardlink.  Thus the disk space is not wasted.

The tool can be run as:
$ fotki --root ROOTDIR --scan SCANDIR [-v] [-n]

where the ROOTDIR is the root of the album, and the SCANDIR is
the directory to start scanning.
