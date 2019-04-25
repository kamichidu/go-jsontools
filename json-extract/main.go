package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/PaesslerAG/jsonpath"
)

func usageFunc(flgs *flag.FlagSet) func() {
	return func() {
		w := flgs.Output()
		fmt.Fprintln(w, "Usage:")
		fmt.Fprintln(w, "  "+flgs.Name()+" [options] {file} {jsonpath}")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options:")
		flgs.PrintDefaults()
	}
}

func run(in io.Reader, out io.Writer, errOut io.Writer, args []string) int {
	log.SetFlags(0)
	log.SetOutput(errOut)

	flgs := flag.NewFlagSet("json-extract", flag.ExitOnError)
	flgs.Usage = usageFunc(flgs)
	if err := flgs.Parse(args[1:]); err != nil {
		log.Print(err)
		return 128
	}
	inName := flgs.Arg(0)
	expr := flgs.Arg(1)

	extract, err := jsonpath.New(expr)
	if err != nil {
		log.Print(err)
		return 128
	}

	var r *bufio.Reader
	if inName == "-" {
		r = bufio.NewReader(in)
	} else {
		file, err := os.Open(inName)
		if err != nil {
			log.Print(err)
			return 1
		}
		defer file.Close()
		r = bufio.NewReader(file)
	}

	for {
		chunk, err := r.ReadBytes('\n')
		if err == io.EOF {
			return 0
		} else if err != nil {
			log.Print(err)
			return 1
		}
		var val interface{}
		if err := json.Unmarshal(chunk, &val); err != nil {
			log.Print(err)
			continue
		}
		if v, err := extract(context.Background(), val); err != nil {
			log.Print(err)
		} else {
			fmt.Fprintf(out, "%v\n", v)
		}
	}
	return 0
}

func main() {
	os.Exit(run(os.Stdin, os.Stdout, os.Stderr, os.Args))
}
