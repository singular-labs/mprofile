# mprofile
GNU Make profiler

This is a simple utility to measure where your GNU makefiles are spending their time building. It works for parallel makes and recursive makes as well. It profiles the execution of your commands and saves the results to a SQLite3 database `mprofile.sqlite` in the top-level directory where you ran `make` from.

## Installing mprofile

You need a reasonably recent Go installed. Build the mprofile binary by running:

```env GOPATH=`pwd` go get -f -t -u -v github.com/singular-labs/mprofile```

(this may take a while, specially in building the go-sqlite3 library).

This will generate the mprofile binary in `./bin/mprofile`. Copy this somewhere in your `$PATH`.

## Using mprofile to generate a profile

Add these lines to the top of your top-level makefile:

```
ifeq ($(DO_MPROFILE), 1)
SHELL= 			mprofile
MPROFILE_TOPDIR=	$(CURDIR)
export MPROFILE_TOPDIR
.SHELLFLAGS= 		T=$@ -c
export DO_MPROFILE
export SHELL
export .SHELLFLAGS
endif
```

and to the top of your recursive makefiles, if any::

```
ifeq ($(DO_MPROFILE), 1)
SHELL = mprofile
.SHELLFLAGS = T=$@ -c
export DO_MPROFILE
endif
```

When you invoke `make DO_MPROFILE=1`, it will generate a profile in `mprofile.sqlite`. The overhead of `mprofile` itself is around 4ms.

mprofile is incremental, i.e. it keeps adding profile data to the SQLite3 database, it doesn't get erased on each successive run. If you want to see only the current run, simply delete `mprofile.sqlite`.

## Database schema

<table>
  <tr><th>Column</th><th>Type</th><th>Description</th></tr>
  <tr><td>level</td><td>integer</td><td>recursive make level (top-level make == 1)</td></tr>
  <tr><td>dir</td><td>text</td><td>working directory of the command</td></tr>
  <tr><td>pid</td><td>integer</td><td>process ID of mprofile</td></tr>
  <tr><td>ppid</td><td>integer</td><td>PID of the make that invoked this target recipe</td></tr>
  <tr><td>p3id</td><td>integer</td><td>PID of the parent of the make, e.g. bash if level&gt;1</td></tr>
  <tr><td>p4id</td><td>integer</td><td>if level&gt;1, PID of the parent mprofile</td></tr>
  <tr><td>p5id</td><td>integer</td><td>if level&gt;1, PID of the parent make</td></tr>
  <tr><td>target</td><td>text</td><td>Makefile target that was being built</td></tr>
  <tr><td>started</td><td>integer</td><td>timestamp of the start of command(s) in nanoseconds since the UNIX epoch (1970-01-01)</td></tr>
  <tr><td>ended</td><td>integer</td><td>timestamp of the end of command(s)</td></tr>
  <tr><td>status</td><td>text</td><td>if not the empty string, the command(s) failed and this is the error message</td></tr>
  <tr><td>cmd</td><td>text</td><td>the command(s) that were requested to run</td></tr>
</table>
