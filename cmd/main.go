// TODO: FLAG, this should no longer be entry point for MIPS
package main

import (
	"errors"
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
	fmt.Printf("Running emu %v\n", len(emu.GBC.Inp.PressedInputs))
	for {
		if err := emu.Update(); err != nil {
			return err
		}
	}
}

// Run program
func Run() int {
	emu := emulator.New()

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
