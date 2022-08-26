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

const (
	ExitCodeOK int = iota
	ExitCodeError
)

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
	flag.Parse()

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
