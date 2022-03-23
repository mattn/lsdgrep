package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	isatty "github.com/mattn/go-isatty"
	lsd "github.com/mattn/go-lsd"
	unicodeclass "github.com/mattn/go-unicodeclass"
)

func main() {
	var distance int
	flag.IntVar(&distance, "d", 2, "distance")
	flag.Parse()

	if flag.NArg() == 0 || flag.NArg() > 2 {
		flag.Usage()
		os.Exit(2)
	}

	var out io.Writer
	var in io.Reader
	var file string

	if isatty.IsTerminal(os.Stdout.Fd()) {
		out = colorable.NewColorableStdout()
	} else {
		out = os.Stdout
	}

	if flag.NArg() == 1 {
		file = "stdin"
		in = os.Stdin
	} else {
		file = flag.Arg(1)
		var err error
		f, err := os.Open(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
			os.Exit(1)
		}
		defer f.Close()
		in = f
	}

	type token struct {
		p int
		l int
	}
	words := unicodeclass.Split(strings.ToLower(flag.Arg(0)))
	scan := bufio.NewScanner(in)
	lno := 0
	for scan.Scan() {
		lno++
		line := scan.Text()
		linewords := unicodeclass.Split(scan.Text())
		if isatty.IsTerminal(os.Stdout.Fd()) {
			tokens := []token{}
			for i := 0; i < len(linewords); i++ {
				if i+len(words) >= len(linewords) {
					break
				}
				js := []int{}
				for j := 0; j < len(words); j++ {
					d := lsd.StringDistance(words[j], strings.ToLower(linewords[i]))
					if d > distance {
						break
					}
					js = append(js, i)
				}
				if len(js) == len(words) {
					tokens = append(tokens, token{
						p: len(strings.Join(linewords[:js[0]], "")),
						l: len(linewords[i]),
					})
				}
			}
			if len(tokens) > 0 {
				fmt.Fprintf(out, "%s:%d:",
					file,
					lno)
				pos := 0
				for _, token := range tokens {
					fmt.Fprint(out, line[pos:token.p])
					fmt.Fprint(out, color.RedString(line[token.p:token.p+token.l]))
					pos = token.p + token.l
				}
				fmt.Fprintln(out, line[pos:])
			}
		} else {
			for i := 0; i < len(linewords); i++ {
				if i+len(words) >= len(linewords) {
					break
				}
				found := 0
				for j := 0; j < len(words); j++ {
					d := lsd.StringDistance(words[j], strings.ToLower(linewords[i+found]))
					if d > distance {
						break
					}
					found++
				}
				if found == len(words) {
					fmt.Fprintf(out, "%s:%d:%s\n",
						file,
						lno,
						strings.Join(linewords, ""))
					break
				}
			}
		}
	}
	if err := scan.Err(); err != nil {
		log.Fatal(err)
	}
}
