package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"

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
	threads    int
}

func init() {
	flag.BoolVar(&cfg.verbose, "v", false, "verbose output")
	flag.StringVar(&cfg.prefix, "p", "", "prefix for command echo")
	flag.StringVar(&cfg.hosts, "hosts", "", "list of hosts")
	flag.StringVar(&cfg.dns, "d", "", "dns name for multi-hosts")
	flag.IntVar(&cfg.threads, "n", 512, "threads to run in parallel")
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

// do runs cmd on host, writing its output to out.
func do(host, cmd string) ([]byte, error) {
	c := exec.Command("ssh", host, cmd)
	output, err := c.CombinedOutput()
	var buf bytes.Buffer

	if err != nil {
		fmt.Fprintf(&buf, "%s[%s]$ %s: Error: %s\n", cfg.prefix, color(host, red, true), cmd, err)
		if len(output) > 0 {
			buf.Write(output)
		}
		return buf.Bytes(), err
	}
	fmt.Fprintf(&buf, "%s[%s]$ %s\n%s", cfg.prefix, color(host, green, true), cmd, string(output))
	return buf.Bytes(), nil
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

	par := cfg.threads
	if par > len(hosts) {
		par = len(hosts)
	}

	// output and input channels
	output := make(chan string, par)
	hostch := make(chan string, par)
	var code int64

	// use par as breadth of parallelism
	var wg, outwg sync.WaitGroup
	wg.Add(par)

	for i := 0; i < par; i++ {
		go func() {
			for host := range hostch {
				out, err := do(host, cmd)
				output <- string(out)
				if err != nil {
					atomic.StoreInt64(&code, 1)
				}
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(output)
	}()

	// print output as it comes in
	outwg.Add(1)
	go func() {
		for o := range output {
			fmt.Print(o)
		}
		outwg.Done()
	}()

	for _, host := range hosts {
		hostch <- host
	}
	close(hostch)
	outwg.Wait()

	os.Exit(int(code))
}
