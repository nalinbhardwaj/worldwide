package emulator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/pokemium/worldwide/pkg/gbc"
	"github.com/pokemium/worldwide/pkg/gbc/cart"
)

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
					return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
					return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
					return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
					return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func Shellout(command string) (error, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return err, stdout.String(), stderr.String()
}


// GameBoy save data is SRAM core dump
func (e *Emulator) writeSav(prevTxNumber int, currentTxNumber int, allPressedInputs []gbc.FrameInput) {
	// Copy to cannon base folder
	os.MkdirAll("cannon/"+strconv.Itoa(currentTxNumber)+"/", os.ModePerm)

	if prevTxNumber > 0 {
		copy("cannon/"+strconv.Itoa(prevTxNumber)+"/"+e.GBC.Cartridge.Title+".sav", "cannon/"+strconv.Itoa(currentTxNumber)+"/"+e.GBC.Cartridge.Title+"-previous.sav")
	}
	copy("pokemon.gbc", "cannon/"+strconv.Itoa(currentTxNumber)+"/"+e.GBC.Cartridge.Title+".gbc")

	savname := "cannon/"+strconv.Itoa(currentTxNumber)+"/"+e.GBC.Cartridge.Title+".sav"

	savfile, err := os.Create(savname) // TODO: FLAG, change for embedded MIPS to output hash
	if err != nil {
		panic(err)
	}
	defer savfile.Close()

	fmt.Printf("writing file: %s\n", savname)

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

	inpname :=  "cannon/"+strconv.Itoa(currentTxNumber)+"/"+e.GBC.Cartridge.Title+".inp.json"

	inpfile, err := os.Create(inpname) // TODO: FLAG, change for embedded MIPS to output hash
	if err != nil {
		panic(err)
	}
	defer inpfile.Close()

	newInp := gbc.Inputs{
		TxNumber: 0,
		PressedInputs:   allPressedInputs,
	}

	fmt.Printf("written pressedinputs: %v\n", len(newInp.PressedInputs))
	inpdata, _ := json.MarshalIndent(newInp, "", " ")

	_, err = inpfile.Write(inpdata)
	if err != nil {
		panic(err)
	}

	// exec cannon with these inputs
	if prevTxNumber > 0 {
		go func() {
			err, out, errout := Shellout("BASEDIR="+"/Users/nibnalin/Documents/worldwide-simul/cannon/"+strconv.Itoa(currentTxNumber)+" /Users/nibnalin/Documents/cannon/mipsevm/mipsevm")
			if err != nil {
					log.Printf("error: %v\n", err)
			}
			fmt.Println("--- stdout ---")
			fmt.Println(out)
			fmt.Println("--- stderr ---")
			fmt.Println(errout)
		}()
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