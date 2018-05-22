// generate profiles of Make execution in SQLite format

package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/fazalmajid/gopsutil/process"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	BACKOFF = 25
)

func parent(pid int) int {
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		log.Fatal(err)
	}
	ppid, err := proc.Ppid()
	if err != nil {
		return -1
	}
	return int(ppid)
}

func main() {
	// command-line options
	// mprofile target
	if len(os.Args) < 3 {
		log.Fatal("insufficient number of command-line parameters: ", os.Args)
	}
	target := os.Args[1]
	if target[0] != 'T' || target[1] != '=' {
		log.Fatal("invalid target argument for mprofile, should start with T=: ", target)
	}
	target = target[2:len(target)]
	curdir, err := os.Getwd()
	if err != nil {
		log.Fatal("could not locate current directory: ", err)
	}

	level, err := strconv.Atoi(os.Getenv("MAKELEVEL"))
	if err != nil {
		log.Fatal("could not identify recursive makefile depth: ", err)
	}
	// log.Println("FFFFFFFFFFFFFFFF MPROFILE: level ", level, " args: ", os.Args)

	// find the parent and grandparent pids so we can build the execution graph
	pid := os.Getpid()  // mprofile's PID
	ppid := parent(pid) // gmake's PID
	p3id := parent(ppid)
	p4id := parent(p3id)
	p5id := parent(p4id)
	// log.Println("mprofile: level", level, pppid, "→", ppid, "→", pid)
	topdir := os.Getenv("MPROFILE_TOPDIR")
	if topdir == "" {
		// log.Println("FFFFFFFFFFFFFFFF could not find MPROFILE_TOPDIR, using ", curdir)
		topdir = curdir
	}
	fn := fmt.Sprintf("%s/mprofile.sqlite", topdir)
	// log.Println("FFFFFFFFFFFFFFFF writing profile to", fn)
	db, err := sql.Open("sqlite3", fn)
	if err != nil {
		log.Fatalf("ERROR: opening SQLite DB %q, error: %s", fn, err)
	}
	// open the actual shell that will run the commands
	// log.Println("FFFFFFFFFFFFFFFF exec ", os.Args[2:len(os.Args)])
	// XXX bash is not the One True (Bourne) Shell
	shell := exec.Command("/bin/bash", os.Args[2:len(os.Args)]...)
	// if running parallel makes (and in the post-Moore's law era of stagnating
	// single-thread performance, shouldn't everyone?) background tasks that
	// attempt to read from the terminal will be suspended with the message
	// "suspended (tty input)" or similar, so we replace this with /dev/null
	// instead
	if terminal.IsTerminal(0) {
		shell.Stdin, err = os.Open("/dev/null")
		if err != nil {
			log.Fatal("Could not open /dev/null for stdin: ", err)
		}
	} else {
		shell.Stdin = os.Stdin
	}
	shell.Stdout = os.Stdout
	shell.Stderr = os.Stderr
	before := time.Now()
	err = shell.Run()
	after := time.Now()
	var status string
	if err != nil {
		status = err.Error()
	}
	var cmdline string
	if os.Args[2] == "-c" && len(os.Args) > 3 {
		cmdline = strings.Join(os.Args[3:len(os.Args)], " ")
	} else {
		cmdline = strings.Join(os.Args[2:len(os.Args)], " ")
	}
	// constructions like $(shell ...) do not pass a target, nor do we really
	// need to profile them
	// if target == "" {
	// 	return
	// }

	// create the schema as needed, we may need to spin if another process has
	// the DB open at the same time
	for {
		_, err = db.Exec("CREATE TABLE IF NOT EXISTS profile (level INTEGER, dir TEXT, pid INTEGER, ppid INTEGER, p3id INTEGER, p4id INTEGER, p5id INTEGER, target TEXT, started INTEGER, ended INTEGER, status TEXT, cmd TEXT)")
		if err != nil {
			if err.Error() != "unable to open database file" {
				log.Fatal("Could not create search table: ", err)
			}
			time.Sleep(time.Duration(rand.Intn(BACKOFF) * 1000000))
		}
		break
	}
	for {
		_, err = db.Exec("INSERT INTO profile (level, dir, pid, ppid, p3id, p4id, p5id, target, started, ended, status, cmd) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", level, curdir, pid, ppid, p3id, p4id, p5id, target, before.UnixNano(), after.UnixNano(), status, cmdline)
		if err != nil {
			log.Fatal("Could not run insert statement: ", err)
		}
		break
	}
	db.Close()
}
