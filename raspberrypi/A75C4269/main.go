package main

import (
	"errors"
	"github.com/djthorpe/gopi"
	_ "github.com/djthorpe/gopi-hw/sys/lirc"
	_ "github.com/djthorpe/gopi/sys/logger"
	"os"
)

const (
	T           = 445
	T8          = T * 8
	T4          = T * 4
	T3          = T * 3
	TracerSpace = T * 20
)

var (
	CallSign = []byte{0x02, 0x20, 0x0E, 0x04, 0x00, 0x00, 0x00, 0x06}
)

func main() {
	config := gopi.NewAppConfig("lirc")
	config.AppFlags.FlagBool("on", false, "power on")
	config.AppFlags.FlagString("mode", "cooler", "cooler or heater or dehumidifier")
	config.AppFlags.FlagUint("t", 28, "16-30 ℃")

	os.Exit(gopi.CommandLineTool(config, func(app *gopi.AppInstance, done chan<- struct{}) error {
		if app.LIRC == nil {
			return errors.New("missing LIRC module")
		}

		var d []uint32
		d = append(d, appendTracerSpace(convertRawSignal(CallSign))...)

		r := Remote{}

		on, _ := app.AppFlags.GetBool("on")
		if on {
			r.Power = PowerOn
		} else {
			r.Power = PowerOff
		}

		mode, _ := app.AppFlags.GetString("mode")
		switch mode {
		case "cooler":
			r.Mode = ModeCooler
		case "heater":
			r.Mode = ModeHeater
		case "dehumidifier":
			r.Mode = ModeDehumidifier
		default:
			r.Mode = ModeCooler
		}

		t, _ := app.AppFlags.GetUint("t")
		r.PresetTemp = t

		d = append(d, convertRawSignal(r.GetSignalBytes())...)

		if err := app.LIRC.PulseSend(d); err != nil {
			return err
		}

		done <- gopi.DONE
		return nil
	}))
}

func convertRawSignal(bytes []byte) []uint32 {
	var seq []uint32

	// Leader
	seq = append(seq, T8, T4)

	// Customer Code 1, 2
	seq = appendByte(seq, bytes[0])
	seq = appendByte(seq, bytes[1])

	// Parity & Data0
	var i uint
	for i = 5; i < 8; i++ {
		if refBit(bytes[2], i) == 1 {
			seq = appendBit1(seq)
		} else {
			seq = appendBit0(seq)
		}
	}
	for i = 0; i < 5; i++ {
		if refBit(bytes[2], i) == 1 {
			seq = appendBit1(seq)
		} else {
			seq = appendBit0(seq)
		}
	}

	// DataN
	for i := 3; i < len(bytes); i++ {
		seq = appendByte(seq, bytes[i])
	}

	// Tracer
	seq = append(seq, T)

	return seq
}

func appendTracerSpace(seq []uint32) []uint32 {
	return append(seq, TracerSpace)
}

func appendBit0(seq []uint32) []uint32 {
	return append(seq, T, T)
}

func appendBit1(seq []uint32) []uint32 {
	return append(seq, T, T3)
}

func refBit(i byte, b uint) byte {
	return (i >> b) & 1
}

func appendByte(seq []uint32, b byte) []uint32 {
	var i uint
	for i = 0; i < 8; i++ {
		if refBit(b, i) == 1 {
			seq = appendBit1(seq)
		} else {
			seq = appendBit0(seq)
		}
	}
	return seq
}
