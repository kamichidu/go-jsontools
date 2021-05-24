package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

func usageFunc(flgs *flag.FlagSet) func() {
	return func() {
		w := flgs.Output()
		fmt.Fprintln(w, "Usage:")
		fmt.Fprintln(w, "  "+flgs.Name()+" [options]")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options:")
		flgs.PrintDefaults()
	}
}

func run(in io.Reader, out io.Writer, errOut io.Writer, args []string) int {
	log.SetFlags(0)
	log.SetOutput(errOut)

	var (
		doDecode bool
	)
	fs := flag.NewFlagSet("json-keycompress", flag.ExitOnError)
	fs.BoolVar(&doDecode, "d", doDecode, "")
	fs.Usage = usageFunc(fs)
	if err := fs.Parse(args[1:]); err != nil {
		log.Print(err)
		return 128
	}

	buffer := json.RawMessage{}
	if err := json.NewDecoder(in).Decode(&buffer); err != nil {
		log.Print(err)
		return 1
	}
	var err error
	if doDecode {
		var dec Decoder
		buffer, err = dec.Decode(buffer)
	} else {
		var enc Encoder
		buffer, err = enc.Encode(buffer)
	}
	if err != nil {
		log.Print(err)
		return 1
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(buffer); err != nil {
		log.Print(err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(os.Stdin, os.Stdout, os.Stderr, os.Args))
}
