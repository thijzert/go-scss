package main

import (
	"flag"
	"fmt"
	"github.com/thijzert/go-scss"
	tc "github.com/thijzert/go-termcolours"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

var (
	act_compile = flag.Bool("compile", false, "Compile source files or directories")
)

func init() {
	flag.Parse()

	// TODO: implement other actions, e.g. "--clean", "--watch", etc.
	if !*act_compile {
		*act_compile = true
	}
}

func main() {
	if *act_compile {
		for _, a := range flag.Args() {
			aa := strings.Split(a, ":")
			source := aa[0]
			target := aa[0]
			if len(aa) > 1 {
				target = aa[1]
			}

			err := compileTo(source, target)

			if err != nil {
				log.Print(err)
				continue
			}
		}
	}
}

func descendInto(sourceDir, targetDir string) error {
	d, err := os.Open(sourceDir)
	if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		if name[0] == '.' {
			continue
		}

		err := compileTo(path.Join(sourceDir, name), path.Join(targetDir, name))
		if err != nil {
			log.Print(err)
		}
	}

	return nil
}

func compileTo(source, target string) error {
	sinf, err := os.Stat(source)
	if err != nil {
		return err
	}
	if sinf.IsDir() {
		return descendInto(source, target)
	}

	// Add the .css suffix
	if len(target) > 5 && target[len(target)-5:] == ".scss" {
		target = target[:len(target)-5] + ".css"
	} else if len(target) > 4 && target[len(target)-4:] == ".css" {
		target = target + ".css"
	}

	nf, err := os.Create(target)
	if err != nil {
		return err
	}
	defer nf.Close()

	src, err := ioutil.ReadFile(source)
	if err != nil {
		return err
	}

	cmp, rerr := scss.Compile(string(src))

	i := 0
	for i < len(cmp) {
		n, err := nf.WriteString(cmp[i:])
		if err != nil {
			return err
		}
		i += n
	}

	if rerr == nil {
		fmt.Printf("    %s %s\n", tc.Green("write"), target)
	}

	return rerr
}
