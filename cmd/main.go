// TODO: FLAG, this should no longer be entry point for MIPS
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pokemium/worldwide/pkg/emulator"
)

var version string

const (
	title = "worldwide"
)

const (
	ExitCodeOK int = iota
	ExitCodeError
)

func init() {
	if version == "" {
		version = "Develop"
	}

	flag.Usage = func() {
		usage := fmt.Sprintf(`Usage:
    %s [arg] [input]
    e.g. %s -p 8888 ./PM_PRISM.gbc
Input: ROM filepath, ***.gb or ***.gbc
Arguments: 
`, title, title)
		fmt.Println(Version())
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
	}
}

func main() {
	os.Exit(Run())
}

func RunGame(emu *emulator.Emulator) error {
	for i := 0; i < emu.GBC.Inp.ExitFrame; i++{
		if err := emu.Update(); err != nil {
			return err
		}
	}
	return nil
}

// Run program
func Run() int {
	var (
		showVersion = flag.Bool("v", false, "show version")
	)

	flag.Parse()

	if *showVersion {
		fmt.Println(Version())
		return ExitCodeOK
	}

	romPath := flag.Arg(0)
	cur, _ := os.Getwd()

	romDir := filepath.Dir(romPath)
	romData, err := readROM(romPath) // TODO: This becomes part of input state in MIPS and gets hashed
	if err != nil {
		fmt.Fprintf(os.Stderr, "ROM Error: %s\n", err)
		return ExitCodeError
	}

	emu := emulator.New(romData, romDir)

	os.Chdir(cur)
	defer func() {
		os.Chdir(cur)
	}()

	if err := RunGame(emu); err != nil {
		if err.Error() == "quit" {
			emu.Exit()
			return ExitCodeOK
		}
		return ExitCodeError
	}
	emu.Exit()
	return ExitCodeOK
}

func Version() string {
	return fmt.Sprintf("%s: %s", title, version)
}

func readROM(path string) ([]byte, error) {
	if path == "" {
		return []byte{}, errors.New("please type .gb or .gbc file path")
	}
	if filepath.Ext(path) != ".gb" && filepath.Ext(path) != ".gbc" {
		return []byte{}, errors.New("please type .gb or .gbc file")
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return []byte{}, errors.New("fail to read file")
	}
	return bytes, nil
}
