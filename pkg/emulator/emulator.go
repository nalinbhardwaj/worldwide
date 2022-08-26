package emulator

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"os"

	"github.com/pokemium/worldwide/geth/common"
	"github.com/pokemium/worldwide/geth/oracle"
	"github.com/pokemium/worldwide/pkg/emulator/audio"
	"github.com/pokemium/worldwide/pkg/emulator/joypad"
	"github.com/pokemium/worldwide/pkg/gbc"
	"github.com/pokemium/worldwide/pkg/gbc/framecountertime"
)

var (
	cache  []byte
)

type Emulator struct {
	GBC      *gbc.GBC
	Rom      []byte
	pause    bool
	reset    bool
	quit     bool
}

func loadOracleData() (common.Hash, *gbc.Inputs, []byte) {
	inputHash := oracle.InputHash()
	fmt.Println("inputHash:", inputHash)
	inputPreimageBytes := oracle.Preimage(inputHash)

	savHash := common.BytesToHash(inputPreimageBytes[0:0x20])
	fmt.Println("savHash:", savHash)
	inpHash := common.BytesToHash(inputPreimageBytes[0x20:0x40])
	fmt.Println("inpHash:", inpHash)
	romHash := common.BytesToHash(inputPreimageBytes[0x40:0x60])
	fmt.Println("romHash:", romHash)

	inpdata := oracle.Preimage(inpHash)
	inpbuf := bytes.NewBuffer(inpdata)
	decoder := gob.NewDecoder(inpbuf)

	var inputs *gbc.Inputs

  decoder.Decode(&inputs)

	fmt.Printf("inpdata loaded %v\n", inputs)

	rom := oracle.Preimage(romHash)
	
	fmt.Printf("romdata loaded %v\n", len(rom))
	return savHash, inputs, rom
}

func New() *Emulator {
	_, inputs, rom := loadOracleData()
	g := gbc.New(rom, joypad.Handler, inputs, audio.SetStream)
	audio.Reset(&g.Sound.Enable)
	e := &Emulator{
		GBC:    g,
		Rom:    rom,
	}
	fmt.Printf("emulator created %v\n", e.GBC.Inp.ExitFrame)

	// e.loadSav(savHash)
	fmt.Printf("save loaded %v\n", e.GBC.Inp.ExitFrame)

	return e
}

func (e *Emulator) Update() error {
	if e.quit {
		return errors.New("quit")
	}
	if e.pause {
		return nil
	}

	defer e.GBC.PanicHandler("update", true)
	shouldStop := e.GBC.Update()
	if shouldStop {
		e.Exit()
		os.Exit(0)
	}
	if e.pause {
		return nil
	}

	audio.Play()

	select {
	case <-framecountertime.Ticker:
		e.GBC.RTC.IncrementSecond()
	default:
	}

	return nil
}

func (e *Emulator) Draw() {
	if e.pause {
		return
	}

	defer e.GBC.PanicHandler("draw", true)
	// cache = e.GBC.Draw()
}

func (e *Emulator) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 160, 144
}

func (e *Emulator) Exit() {
	e.writeSav()
}

// TODO: remove reset from non MIPS version?