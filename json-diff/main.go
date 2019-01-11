package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	jd "github.com/josephburnett/jd/lib"
	"github.com/mattn/go-colorable"
)

func usageFunc(flgs *flag.FlagSet) func() {
	return func() {
		w := flgs.Output()
		fmt.Fprintln(w, "Usage:")
		fmt.Fprintln(w, "  "+flgs.Name()+" [options] {fileA} {fileB}")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options:")
		flgs.PrintDefaults()
	}
}

func readFile(name string) (os.FileInfo, jd.JsonNode, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return nil, nil, err
	}
	jnode, err := jd.ReadJsonFile(name)
	if err != nil {
		return nil, nil, err
	}
	return stat, jnode, nil
}

func run(in io.Reader, out io.Writer, errOut io.Writer, args []string) int {
	log.SetFlags(0)
	log.SetOutput(errOut)

	fs := flag.NewFlagSet("json_diff", flag.ExitOnError)
	fs.Usage = usageFunc(fs)
	if err := fs.Parse(args[1:]); err != nil {
		log.Print(err)
		return 128
	} else if fs.NArg() != 2 {
		fs.Usage()
		return 128
	}

	aName := fs.Arg(0)
	aStat, aJNode, err := readFile(aName)
	if err != nil {
		log.Print(err)
		return 1
	}
	bName := fs.Arg(1)
	bStat, bJNode, err := readFile(bName)
	if err != nil {
		log.Print(err)
		return 1
	}

	// want "diff -u" output
	const diffTimeLayout = "2006-01-02 15:04:05.000000000 -0700"
	fmt.Fprintf(out, "--- %q\t%s\n", aName, aStat.ModTime().Format(diffTimeLayout))
	fmt.Fprintf(out, "+++ %q\t%s\n", bName, bStat.ModTime().Format(diffTimeLayout))
	fmt.Fprintln(out, aJNode.Diff(bJNode).Render())
	return 0
}

func main() {
	os.Exit(run(os.Stdin, colorable.NewColorable(os.Stdout), colorable.NewColorable(os.Stderr), os.Args))
}
