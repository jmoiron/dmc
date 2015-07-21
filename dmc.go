package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// dmc runs the command on all hosts passed via stdin simultaneously

const (
	white = iota + 89
	black
	red
	green
	yellow
	blue
	purple
)

func color(s string, color int, bold bool) string {
	b := "01;"
	if !bold {
		b = ""
	}
	return fmt.Sprintf("\033[%s%dm%s\033[0m", b, color, s)
}

var cfg struct {
	verbose    bool
	interleave bool
	prefix     string
	hosts      string
}

func init() {
	flag.BoolVar(&cfg.verbose, "v", false, "verbose output")
	flag.StringVar(&cfg.prefix, "p", "", "prefix for command echo")
	flag.StringVar(&cfg.hosts, "hosts", "", "list of hosts")
	// flag.BoolVar(&cfg.interleave, "i", false, "interleave output as it is available")
	flag.Parse()
}

func vprintf(format string, args ...interface{}) {
	if cfg.verbose {
		fmt.Printf(format, args...)
	}
}

func getHosts() []string {
	if len(cfg.hosts) > 0 {
		return strings.Split(cfg.hosts, ",")
	}

	var hosts []string
	fi, _ := os.Stdin.Stat()
	if (fi.Mode() & os.ModeCharDevice) != 0 {
		fmt.Println("usage: you must pipe a list of hosts into dmc or use -hosts.")
		return hosts
	}
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		hosts = append(hosts, strings.Trim(s.Text(), "\n"))
	}
	if err := s.Err(); err != nil {
		fmt.Printf("Error reading from stdin: %s\n", err)
	}
	return hosts

}

func main() {
	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("usage: dmc <command>")
		return
	}

	hosts := getHosts()
	cmd := strings.Join(args, " ")
	vprintf("Running `%s` on %d hosts\n", cmd, len(hosts))

	var wg sync.WaitGroup
	wg.Add(len(hosts))
	output := make(chan string)
	for _, host := range hosts {
		go func(host string) {
			defer wg.Done()
			c := exec.Command("ssh", host, cmd)
			out, err := c.CombinedOutput()
			if err != nil {
				e := fmt.Sprintf("%s[%s]$ %s: Error: %s\n", cfg.prefix, color(host, red, true), cmd, err)
				if len(out) > 0 {
					e = fmt.Sprintf("%s%s", e, string(out))
				}
				output <- e
				return
			}
			output <- fmt.Sprintf("%s[%s]$ %s\n%s", cfg.prefix, color(host, green, true), cmd, string(out))
		}(host)
	}

	go func() {
		wg.Wait()
		close(output)
	}()

	for o := range output {
		fmt.Print(o)
	}

}
