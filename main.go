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
		tokens := []token{}
		pos := 0
		for i := 0; i < len(linewords); i++ {
			if i+len(words) >= len(linewords) {
				break
			}
			found := 0
			total := distance
			for j := 0; j < len(words); j++ {
				total -= lsd.StringDistance(words[j], strings.ToLower(linewords[i+found]))
				found++
			}
			if total >= 0 {
				tokens = append(tokens, token{
					p: pos,
					l: len(strings.Join(linewords[i:i+found], "")),
				})
			}
			pos += len(linewords[i])
		}
		if len(tokens) > 0 {
			fmt.Fprintf(out, "%s:%d:",
				file,
				lno)
			pos := 0
			fmt.Println(tokens)
			for _, token := range tokens {
				fmt.Fprint(out, line[pos:token.p])
				if isatty.IsTerminal(os.Stdout.Fd()) {
					fmt.Fprint(out, color.RedString(line[token.p:token.p+token.l]))
				} else {
					fmt.Fprint(out, line[token.p:token.p+token.l])
				}
				pos = token.p + token.l
			}
			fmt.Fprintln(out, line[pos:])
		}
	}
	if err := scan.Err(); err != nil {
		log.Fatal(err)
	}
}
