package amiga

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	LEN_NAME        = 20
	LEN_SAMPLE_NAME = 22
	NUM_SAMPLES     = 31
	LEN_SEQUENCE    = 128
)

type Sample struct {
	Name        []byte
	Len         uint16
	Tune        int8
	Volume      uint8
	RepeatStart uint16
	RepeatLen   uint16
	Data        []byte
}

type Pattern struct {
	// TODO Change Pattern stuct to store parsed Note data
	Data []byte
}

type Mod struct {
	Name     []byte
	Samples  []Sample
	Len      byte
	Sequence []byte
	Patterns []Pattern
}

func Read(r io.Reader) (Mod, error) {
	var mod Mod

	name, err := readName(r, LEN_NAME)
	if err != nil {
		return Mod{}, err
	}
	mod.Name = name

	// TODO read correct number of samples
	for i := 0; i < NUM_SAMPLES; i++ {
		sample, err := readSample(r)
		if err != nil {
			return Mod{}, err
		}
		mod.Samples = append(mod.Samples, sample)
	}

	len, err := readByte(r)
	if err != nil {
		return Mod{}, err
	}
	// TODO check range [1,128]
	mod.Len = len

	_, err = readByte(r)
	if err != nil {
		return Mod{}, err
	}
	// TODO check =127

	seq, err := readBytes(r, LEN_SEQUENCE)
	if err != nil {
		return Mod{}, err
	}
	// TODO check values [0,63]
	// TODO remove things past len?
	mod.Sequence = seq

	mk, err := readBytes(r, 4)
	if err != nil {
		return Mod{}, err
	}
	if string(mk) != "M.K." {
		return Mod{}, errors.New("Not M.K.")
	}

	var npatterns byte
	for _, patternid := range seq {
		if patternid > npatterns {
			npatterns = patternid
		}
	}
	for i := byte(0); i < npatterns; i++ {
		pattern, err := readPattern(r)
		if err != nil {
			return Mod{}, nil
		}
		mod.Patterns = append(mod.Patterns, pattern)
	}

	for i := 0; i < NUM_SAMPLES; i++ {
		if len := mod.Samples[i].Len; len > 0 {
			data, err := readBytes(r, int(len))
			if err != nil {
				return Mod{}, nil
			}
			mod.Samples[i].Data = data
		}
	}

	return mod, nil
}

func readSample(r io.Reader) (Sample, error) {
	var sample Sample

	name, err := readName(r, LEN_SAMPLE_NAME)
	if err != nil {
		return sample, err
	}
	sample.Name = name

	len, err := readWyde(r)
	if err != nil {
		return sample, err
	}
	sample.Len = len * 2

	tune, err := readNibble(r)
	if err != nil {
		return sample, err
	}
	sample.Tune = tune

	vol, err := readByte(r)
	if err != nil {
		return sample, err
	}
	// TODO check 0<=vol<=64
	sample.Volume = vol

	rstart, err := readWyde(r)
	if err != nil {
		return sample, err
	}
	sample.RepeatStart = rstart * 2

	rlen, err := readWyde(r)
	if err != nil {
		return sample, err
	}
	sample.RepeatLen = rlen * 2

	return sample, nil
}

func readPattern(r io.Reader) (Pattern, error) {
	data, err := readBytes(r, 1024)
	if err != nil {
		return Pattern{}, err
	}
	return Pattern{data}, nil
}

func readName(r io.Reader, len int) ([]byte, error) {
	name, err := readBytes(r, len)
	if err != nil {
		return nil, err
	}
	for i, b := range name {
		if b == 0 {
			name = name[:i]
			break
		}
	}
	return name, nil
}

func readWyde(r io.Reader) (uint16, error) {
	b, err := readBytes(r, 2)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(b), nil
}

func readNibble(r io.Reader) (int8, error) {
	b, err := readBytes(r, 1)
	if err != nil {
		return 0, err
	}
	return int8(b[0] & 0x0F), nil
}

func readByte(r io.Reader) (byte, error) {
	b, err := readBytes(r, 1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

func readBytes(r io.Reader, len int) ([]byte, error) {
	b := make([]byte, len)
	_, err := io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
