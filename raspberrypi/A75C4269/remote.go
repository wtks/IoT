package main

const (
	PowerOff byte = iota
	PowerOn
	PowerOnAndOffTimer
	PowerOffAndOnTimer

	ModeCooler byte = iota
	ModeHeater
	ModeDehumidifier

	DirectionAuto byte = iota
	Direction1
	Direction2
	Direction3
	Direction4
	Direction5

	AirVolumeAuto byte = iota
	AirVolumeStill
	AirVolume1
	AirVolume2
	AirVolume3
	AirVolume4
	AirVolumePowerful
)

type Remote struct {
	Power         byte
	Mode          byte
	PresetTemp    uint
	AirVolume     byte
	WindDirection byte
	TimerHour     byte
}

func (r *Remote) GetSignalBytes() (result []byte) {
	var b byte

	// 1-6byte: header
	result = append(result, 0x02, 0x20, 0x00, 0x0E, 0x04, 0x00)

	// 7byte: mode, timer, power
	b = 0
	switch r.Mode {
	case ModeCooler:
		b |= 0x3 // 0011
	case ModeHeater:
		b |= 0x4 // 0100
	case ModeDehumidifier:
		b |= 0x2 // 0010
	default:
		b |= 0x3 // Cooler
	}
	b <<= 4
	switch r.Power {
	case PowerOff:
		b |= 0x0 // 0000
	case PowerOn:
		b |= 0x1 // 0001
	case PowerOnAndOffTimer:
		b |= 0x5 // 0101
	case PowerOffAndOnTimer:
		b |= 0x2 // 0010
	default:
		b |= 0x0 // PowerOff
	}
	result = append(result, b)

	// 8byte: temp
	b = 0x20 // 00100000
	switch {
	case r.PresetTemp < 16:
		b |= 0x0 << 1 // 0000
	case r.PresetTemp > 30:
		b |= 0xE << 1 // 1110
	default:
		b |= byte(r.PresetTemp-16) << 1
	}
	result = append(result, b)

	// 9byte: ?
	b = 0x80
	result = append(result, b)

	// 10byte: air volume, wind direction
	switch r.AirVolume {
	case AirVolumeAuto:
		b = 0xA << 4 // 1010
	case AirVolume1, AirVolumeStill, AirVolumePowerful:
		b = 0x3 << 4 // 0011
	case AirVolume2:
		b = 0x4 << 4 // 0101
	case AirVolume3:
		b = 0x5 << 4 // 0110
	case AirVolume4:
		b = 0x6 << 4 // 0111
	default:
		b = 0xA << 4 // Auto
	}
	switch r.WindDirection {
	case DirectionAuto:
		b |= 0xF // 1111
	case Direction1:
		b |= 0x1 // 0001
	case Direction2:
		b |= 0x2 // 0010
	case Direction3:
		b |= 0x3 // 0011
	case Direction4:
		b |= 0x4 // 0100
	case Direction5:
		b |= 0x5 // 0101
	default:
		b |= 0xF // Auto
	}
	result = append(result, b)

	// 11byte: ?
	b = 0x00
	result = append(result, b)

	// 12byte: timer setting
	b = 0
	if r.Power == PowerOffAndOnTimer || r.Power == PowerOnAndOffTimer {
		b = 0x3C
	}
	result = append(result, b)

	// 13-14byte: timer hour
	switch r.TimerHour {
	case 1:
		result = append(result, 0xC0, 0x03) // 11000000 00000011
	case 2:
		result = append(result, 0x80, 0x07) // 10000000 00000111
	case 3:
		result = append(result, 0x40, 0x0B) // 01000000 00001011
	case 4:
		result = append(result, 0x00, 0x0F) // 00000000 00001111
	case 5:
		result = append(result, 0xC0, 0x12) // 11000000 00010010
	case 6:
		result = append(result, 0x80, 0x16) // 10000000 00010110
	case 7:
		result = append(result, 0x40, 0x1A) // 01000000 00011010
	case 8:
		result = append(result, 0x00, 0x1E) // 00000000 00011110
	case 9:
		result = append(result, 0xC0, 0x21) // 11000000 00100001
	case 10:
		result = append(result, 0x80, 0x25) // 10000000 00100101
	case 11:
		result = append(result, 0x40, 0x29) // 01000000 00101001
	case 12:
		result = append(result, 0x00, 0x2D) // 00000000 00101101
	default:
		result = append(result, 0xC0, 0x03) // 1 hour
	}

	// 15byte: Various Flags
	b = 0
	if r.AirVolume == AirVolumeStill {
		b |= 1
	}
	b <<= 5
	if r.AirVolume == AirVolumePowerful {
		b |= 1
	}
	result = append(result, b)

	// 16byte: Various Flags
	b = 0
	if r.PresetTemp <= 16 || r.PresetTemp >= 30 {
		b |= 1
	}
	b <<= 1
	result = append(result, b)

	// 17byte: ?
	b = 0x80
	result = append(result, b)

	// 18byte: ?
	b = 0x00
	result = append(result, b)

	// 19byte: ?
	b = 0x06
	result = append(result, b)

	// 20byte: checksum
	sum := 0x6
	for i := 6; i < 19; i++ {
		sum += int(result[i])
	}
	b = byte(0xFF & sum)
	result = append(result, b)

	return result
}