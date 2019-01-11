package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/mattn/go-colorable"
	"gopkg.in/yaml.v2"
)

type formatList []string

func (l formatList) IsValidFormat(s string) bool {
	for _, el := range l {
		if el == s {
			return true
		}
	}
	return false
}

func (l formatList) String() string {
	v := make([]string, len(l))
	for i := range l {
		v[i] = fmt.Sprintf("%q", l[i])
	}
	return strings.Join(v, ", ")
}

var (
	availableFormats = formatList{"yaml"}
)

func usageFunc(flgs *flag.FlagSet) func() {
	return func() {
		w := flgs.Output()
		fmt.Fprintln(w, "Usage:")
		fmt.Fprintln(w, "  "+flgs.Name()+" [options] {file}")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options:")
		flgs.PrintDefaults()
	}
}

func readYAML(r io.Reader) (interface{}, error) {
	yd := yaml.NewDecoder(r)
	var v interface{}
	if err := yd.Decode(&v); err != nil {
		return nil, err
	}
	return toJSONCompatType(v), nil
}

func toJSONCompatType(v interface{}) interface{} {
	switch val := v.(type) {
	case []interface{}:
		for i := range val {
			val[i] = toJSONCompatType(val[i])
		}
		return val
	case map[interface{}]interface{}:
		out := map[string]interface{}{}
		for k, v := range val {
			out[fmt.Sprint(k)] = toJSONCompatType(v)
		}
		return out
	default:
		return v
	}
}

func run(in io.Reader, out io.Writer, errOut io.Writer, args []string) int {
	log.SetFlags(0)
	log.SetOutput(errOut)

	fs := flag.NewFlagSet("json_from", flag.ExitOnError)
	fs.Usage = usageFunc(fs)
	format := fs.String("format", "yaml", fmt.Sprintf("read file as this format (choices: %s)", availableFormats.String()))
	pretty := fs.Bool("pretty", false, "pretty json output")
	if err := fs.Parse(args[1:]); err != nil {
		log.Print(err)
		return 128
	} else if fs.NArg() != 1 {
		fs.Usage()
		return 128
	} else if !availableFormats.IsValidFormat(*format) {
		fs.Usage()
		return 128
	}

	var r io.Reader
	if filename := fs.Arg(0); filename == "-" {
		r = in
	} else {
		file, err := os.Open(filename)
		if err != nil {
			log.Print(err)
			return 1
		}
		defer file.Close()
		r = file
	}
	r = bufio.NewReader(r)

	var v interface{}
	var err error
	switch *format {
	case "yaml":
		v, err = readYAML(r)
	}
	if err != nil {
		log.Print(err)
		return 1
	}
	out = bufio.NewWriter(out)
	defer func() {
		if flusher, ok := out.(*bufio.Writer); ok {
			flusher.Flush()
		}
	}()
	je := json.NewEncoder(out)
	if *pretty {
		je.SetIndent("", "  ")
	}
	if err := je.Encode(v); err != nil {
		log.Print(err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(os.Stdin, colorable.NewColorable(os.Stdout), colorable.NewColorable(os.Stderr), os.Args))
}
