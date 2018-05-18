// generate profiles of Make execution in SQLite format

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/fazalmajid/gopsutil/process"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// command-line options
	// mprofile target
	if len(os.Args) < 3 {
		log.Fatal("insufficient number of command-line parameters", os.Args)
	}
	target := os.Args[1]
	curdir := os.Args[2]

	level, err := strconv.Atoi(os.Getenv("MAKELEVEL"))
	if err != nil {
		log.Fatal("could not identify recursive makefile depth:", err)
	}
	// find the parent and grandparent pids so we can build the execution graph
	pid := os.Getpid()
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		log.Fatal(err)
	}
	ppid, err := proc.Ppid()
	if err != nil {
		log.Fatal(err)
	}
	proc, err = process.NewProcess(int32(ppid))
	if err != nil {
		log.Fatal(err)
	}
	pppid, err := proc.Ppid()
	if err != nil {
		log.Fatal(err)
	}
	// log.Println("mprofile: level", level, pppid, "→", ppid, "→", pid)
	fn := fmt.Sprintf("%s/mprofile.sqlite", curdir)
	// log.Println("writing profile to", fn)
	db, err := sql.Open("sqlite3", fn)
	if err != nil {
		log.Fatalf("ERROR: opening SQLite DB %q, error: %s", fn, err)
	}
	// create the schema as needed
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS profile (level INTEGER, dir TEXT, pid INTEGER, ppid INTEGER, pppid INTEGER, target TEXT, started INTEGER, ended INTEGER, status TEXT, cmd TEXT)")
	if err != nil {
		log.Fatal("Could not create search table", err)
	}
	// open the actual shell that will run the commands
	shell := exec.Command("/bin/sh", os.Args[3:len(os.Args)]...)
	shell.Stdin = os.Stdin
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
	if os.Args[3] == "-c" && len(os.Args) > 4 {
		cmdline = strings.Join(os.Args[4:len(os.Args)], " ")
	} else {
		cmdline = strings.Join(os.Args[3:len(os.Args)], " ")
	}

	_, err = db.Exec("INSERT INTO profile (level, dir, pid, ppid, pppid, target, started, ended, status, cmd) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", level, curdir, pid, ppid, pppid, target, before.UnixNano(), after.UnixNano(), status, cmdline)
	if err != nil {
		log.Fatal("Could not run insert statement: ", err)
	}
	db.Close()
}
