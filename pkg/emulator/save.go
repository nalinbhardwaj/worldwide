package emulator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pokemium/worldwide/pkg/gbc/cart"
)

// GameBoy save data is SRAM core dump
func (e *Emulator) writeSav(currentTxNumber int) {
	savname := filepath.Join(e.RomDir, strconv.Itoa(currentTxNumber), "-"+e.GBC.Cartridge.Title+".sav")

	savfile, err := os.Create(savname) // TODO: FLAG, change for embedded MIPS to output hash
	if err != nil {
		return
	}
	defer savfile.Close()

	var buffer []byte
	switch e.GBC.Cartridge.RAMSize {
	case cart.RAM_UNUSED:
		buffer = make([]byte, 0x800)
		for index := 0; index < 0x800; index++ {
			buffer[index] = e.GBC.RAM.Buffer[0][index]
		}
	case cart.RAM_8KB:
		buffer = make([]byte, 0x2000*1)
		for index := 0; index < 0x2000; index++ {
			buffer[index] = e.GBC.RAM.Buffer[0][index]
		}
	case cart.RAM_32KB:
		buffer = make([]byte, 0x2000*4)
		for i := 0; i < 4; i++ {
			for j := 0; j < 0x2000; j++ {
				index := i*0x2000 + j
				buffer[index] = e.GBC.RAM.Buffer[i][j]
			}
		}
	case cart.RAM_64KB:
		buffer = make([]byte, 0x2000*8)
		for i := 0; i < 8; i++ {
			for j := 0; j < 0x2000; j++ {
				index := i*0x2000 + j
				buffer[index] = e.GBC.RAM.Buffer[i][j]
			}
		}
	}

	if e.GBC.Cartridge.HasRTC() {
		rtcData := e.GBC.RTC.Dump()
		for i := 0; i < 48; i++ {
			buffer = append(buffer, rtcData[i])
		}
	}

	fmt.Printf("savdata buffer: %x\n", buffer)

	_, err = savfile.Write(buffer)
	if err != nil {
		panic(err)
	}

	// inpname := filepath.Join(e.RomDir, strconv.Itoa(currentTxNumber), "-"+e.GBC.Cartridge.Title+".inp")

	// inpfile, err := os.Create(inpname) // TODO: FLAG, change for embedded MIPS to output hash
	// if err != nil {
	// 	panic(err)
	// }
	// defer inpfile.Close()

	// encoder := gob.NewEncoder(inpfile)
	// encoder.Encode(e.GBC.PressedInputs)
}

func (e *Emulator) loadSav() {
	savname := filepath.Join(e.RomDir, e.GBC.Cartridge.Title+".sav")

	savdata, err := os.ReadFile(savname) // TODO: FLAG, change for embedded MIPS to input hash, preimage etc
	if err != nil {
		return
	}
	fmt.Printf("savdata: %x\n", savdata)

	switch e.GBC.Cartridge.RAMSize {
	case cart.RAM_UNUSED:
		for index := 0; index < 0x800; index++ {
			e.GBC.RAM.Buffer[0][index] = savdata[index]
		}
	case cart.RAM_8KB:
		for index := 0; index < 0x2000; index++ {
			e.GBC.RAM.Buffer[0][index] = savdata[index]
		}
	case cart.RAM_32KB:
		for i := 0; i < 4; i++ {
			for j := 0; j < 0x2000; j++ {
				index := i*0x2000 + j
				e.GBC.RAM.Buffer[i][j] = savdata[index]
			}
		}
	case cart.RAM_64KB:
		for i := 0; i < 8; i++ {
			for j := 0; j < 0x2000; j++ {
				index := i*0x2000 + j
				e.GBC.RAM.Buffer[i][j] = savdata[index]
			}
		}
	}

	if e.GBC.Cartridge.HasRTC() {
		start := (len(savdata) / 0x1000) * 0x1000
		rtcData := savdata[start : start+48]
		e.GBC.RTC.Sync(rtcData)
	}
}

func (e *Emulator) loadInp(currentTxNumber int) bool {
	inpname := filepath.Join(e.RomDir, e.GBC.Cartridge.Title+"-"+strconv.Itoa(currentTxNumber)+".inp.json")

	file, err := os.ReadFile(inpname)
	if err != nil {
		return false
	}
	_ = json.Unmarshal([]byte(file), &e.GBC.Inp)
	return true
}