package emulator

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hajimehoshi/ebiten/v2"
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
	RomDir   string
	pause    bool
	reset    bool
	quit     bool
}

func New(romData []byte, romDir string) *Emulator {
	g := gbc.New(romData, joypad.Handler, audio.SetStream)
	audio.Reset(&g.Sound.Enable)

	ebiten.SetWindowResizable(true)
	ebiten.SetWindowTitle("60fps")
	ebiten.SetWindowSize(160*2, 144*2)

	e := &Emulator{
		GBC:    g,
		Rom:    romData,
		RomDir: romDir,
	}
	e.setupCloseHandler()

	e.loadSav()
	return e
}

func (e *Emulator) ResetGBC() {
	e.writeSav(currentTxNumber)

	oldCallbacks := e.GBC.Callbacks
	e.GBC = gbc.New(e.Rom, joypad.Handler, audio.SetStream)
	e.GBC.Callbacks = oldCallbacks

	e.loadSav()
	framecountertime.SetUnixNow(1651411507)

	e.reset = false
}

var currentTxNumber = 0
var currentTxIsDone = false
var currentTxInputCounter = 0
const TXBATCHSIZE = 2

func (e *Emulator) loadCurrentTx() bool {
	didRead := e.loadInp(currentTxNumber + 1)
	if didRead {
		currentTxNumber = e.GBC.Inp.TxNumber
		currentTxIsDone = false
		currentTxInputCounter = 0
		fmt.Println("new tx", e.GBC.Inp.TxNumber, len(e.GBC.Inp.PressedInputs))
		return false
	}
	return true
}

func (e *Emulator) Update() error {
	if currentTxIsDone {
		noNewTx := e.loadCurrentTx()
		if noNewTx {
			// Paused
			return nil
		}
	}
	if e.quit {
		return errors.New("quit")
	}
	if e.reset {
		e.ResetGBC()
		return nil
	}
	if e.pause {
		return nil
	}

	defer e.GBC.PanicHandler("update", true)
	currentTxStatus := e.GBC.Update(currentTxInputCounter)
	currentTxInputCounter++
	if currentTxStatus {
		fmt.Println("current tx is done")
		currentTxIsDone = true
	}
	if currentTxIsDone && currentTxNumber % TXBATCHSIZE == 0 && currentTxNumber > 0 {
		// Dump sav and restart
		fmt.Println("current tx is done, restarting")
		e.ResetGBC()
	}
	if e.pause {
		return nil
	}

	audio.Play()

	select {
	case <-framecountertime.Ticker:
		e.GBC.RTC.IncrementSecond()
		ebiten.SetWindowTitle(fmt.Sprintf("%dfps", int(ebiten.CurrentTPS())))
	default:
	}

	return nil
}

func (e *Emulator) Draw(screen *ebiten.Image) {
	if e.pause {
		screen.ReplacePixels(cache)
		return
	}

	defer e.GBC.PanicHandler("draw", true)
	cache = e.GBC.Draw()
	screen.ReplacePixels(cache)
}

func (e *Emulator) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 160, 144
}

func (e *Emulator) Exit() {
	e.writeSav(currentTxNumber)
}

func (e *Emulator) setupCloseHandler() { // TODO: BIG FLAG, need to delete this from MIPS
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		e.Exit()
		os.Exit(0)
	}()
}
