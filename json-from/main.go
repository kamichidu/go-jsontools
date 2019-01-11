package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
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

type Flusher interface {
	Flush() error
}

func replaceExtension(name string, newExt string) string {
	dir, file := filepath.Split(name)
	ext := filepath.Ext(file)
	file = strings.TrimSuffix(file, ext) + newExt
	return filepath.Join(dir, file)
}

func flush(w io.Writer) error {
	flusher, ok := w.(Flusher)
	if !ok {
		return nil
	}
	return flusher.Flush()
}

func run(in io.Reader, out io.Writer, errOut io.Writer, args []string) int {
	log.SetFlags(0)
	log.SetOutput(errOut)

	fs := flag.NewFlagSet("json_from", flag.ExitOnError)
	fs.Usage = usageFunc(fs)
	format := fs.String("format", "yaml", fmt.Sprintf("read file as this format (choices: %s)", availableFormats.String()))
	pretty := fs.Bool("pretty", false, "pretty json output")
	autoOutName := fs.Bool("O", false, "write output to a file named as the input file replacing its extension to .json")
	outName := fs.String("o", "", "write to file instead of stdout")
	if err := fs.Parse(args[1:]); err != nil {
		log.Print(err)
		return 128
	} else if fs.NArg() != 1 {
		fs.Usage()
		return 128
	} else if !availableFormats.IsValidFormat(*format) {
		fs.Usage()
		return 128
	} else if *autoOutName && *outName != "" {
		log.Print("-O and -o are exclusive flag")
		return 128
	}
	inName := fs.Arg(0)

	var r io.Reader
	if inName == "-" {
		r = in
	} else {
		file, err := os.Open(inName)
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

	var w io.Writer
	if *autoOutName {
		*outName = replaceExtension(inName, ".json")
	}
	if *outName != "" {
		file, err := os.Create(*outName)
		if err != nil {
			log.Print(err)
			return 1
		}
		defer file.Close()
		w = file
	} else {
		w = out
	}
	w = bufio.NewWriter(w)
	defer flush(w)

	je := json.NewEncoder(w)
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
