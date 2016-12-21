package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh/terminal"
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

var tty = terminal.IsTerminal(int(os.Stdout.Fd()))

func color(s string, color int, bold bool) string {
	if !tty {
		return s
	}
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
	dns        string
	maxHosts   int
}

func init() {
	flag.BoolVar(&cfg.verbose, "v", false, "verbose output")
	flag.StringVar(&cfg.prefix, "p", "", "prefix for command echo")
	flag.StringVar(&cfg.hosts, "hosts", "", "list of hosts")
	flag.StringVar(&cfg.dns, "d", "", "dns name for multi-hosts")
	flag.IntVar(&cfg.maxHosts, "m", 1024, "Maximum number of hosts the command will run on in parallel")
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
	if len(cfg.dns) > 0 {
		hosts, err := net.LookupHost(cfg.dns)
		if err != nil {
			fmt.Printf("Error looking up %s: %s\n", cfg.dns, err)
			os.Exit(-1)
		}
		return hosts
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

	groups := len(hosts) / cfg.maxHosts
	if len(hosts) % cfg.maxHosts != 0 {
		groups++
	}

	var wg sync.WaitGroup
	var once sync.Once
	wg.Add(len(hosts))
	output := make(chan string)
	return_code := 0
	setError := func() {
		return_code = 1
	}

	for i := 0; i < groups; i++ {
		beginRange := cfg.maxHosts * i
		endRange := min(cfg.maxHosts * (i + 1), len(hosts))
		for _, host := range hosts[beginRange:endRange] {
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
					once.Do(setError)
					return
				}
				output <- fmt.Sprintf("%s[%s]$ %s\n%s", cfg.prefix, color(host, green, true), cmd, string(out))
			}(host)
		}
	}

	go func() {
		wg.Wait()
		close(output)
	}()

	for o := range output {
		fmt.Print(o)
	}

	os.Exit(return_code)
}
