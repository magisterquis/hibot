/*
 * hibot.go
 * by J. Stuart McMurray
 * quicky bot for shaow's ssh-chat
 * created 20141213
 * last modified 20141213
 */

package main

import (
	"bufio"
	"fmt"
	"net/textproto"
	"os"
	"os/exec"
	"regexp"
	"time"
)

func main() { os.Exit(mymain()) }
func mymain() int {
	/* Connect to the server */
	cmd := exec.Command("ssh", "chat.shazow.net")
	o, err := cmd.StdoutPipe()
	if nil != err {
		fmt.Printf("Unable to get stdout pipe: %v\n", err)
		return -1
	}
	i, err := cmd.StdinPipe()
	if nil != err {
		fmt.Printf("Unable to get stdin pipe: %v\n", err)
		return -2
	}
	bufo := textproto.NewReader(bufio.NewReader(o))
	bufi := textproto.NewWriter(bufio.NewWriter(i))
	cmd.Start()

	/* Set the nick */
	bufi.PrintfLine("/nick hibot")

	/* Goroutine to greet people in a rate-limited fashion */
	hichan := make(chan string)
	go func() {
		for {
			nick := <-hichan
			bufi.PrintfLine("Welcome, %v.\n", nick)
			fmt.Printf("Greeted %v\n", nick)
			time.Sleep(time.Second)
		}
	}()

	/* Eat the first 10 lines, which are history */
	for i := 0; i < 10; i++ {
		if _, err := bufo.ReadLine(); nil != err {
			fmt.Printf("Initial clear error: %v\n", err)
			return -4
		}
	}

	/* Watch for joins */
	r := regexp.MustCompile(`\* (\S+) joined`)
	for {
		/* Get a line */
		line, err := bufo.ReadLine()
		if nil != err {
			fmt.Printf("Read err: %v\n", err)
			break
		}
		fmt.Printf("-> %v\n", line)
		/* Check for a join */
		m := r.FindStringSubmatch(line)
		/* On a join, queue(ish) up a hello */
		if m != nil && len(m) > 1 {
			go func() { hichan <- m[1] }()
		}
	}
	/* Die gracefully */
	fmt.Printf("Done\n")
	i.Close()
	fmt.Printf("Closed processe's stdin\n")
	cmd.Wait()
	fmt.Printf("Process done\n")
	return -3
}
