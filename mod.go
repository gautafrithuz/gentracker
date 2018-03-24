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

func (m *Mod) Read(r io.Reader) error {
	name, err := readName(r, LEN_NAME)
	if err != nil {
		return err
	}
	m.Name = name

	// TODO read correct number of samples based on initials
	for i := 0; i < NUM_SAMPLES; i++ {
		sample, err := readSampleHead(r)
		if err != nil {
			return err
		}
		m.Samples = append(m.Samples, sample)
	}

	len, err := readByte(r)
	if err != nil {
		return err
	}
	// TODO check range [1,128]
	m.Len = len

	_, err = readByte(r)
	if err != nil {
		return err
	}
	// TODO check =127

	seq, err := readBytes(r, LEN_SEQUENCE)
	if err != nil {
		return err
	}
	// TODO check values [0,63]
	m.Sequence = seq

	mk, err := readBytes(r, 4)
	if err != nil {
		return err
	}
	if string(mk) != "M.K." {
		return errors.New("Not M.K.")
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
			return err
		}
		m.Patterns = append(m.Patterns, pattern)
	}

	for i := 0; i < NUM_SAMPLES; i++ {
		if len := m.Samples[i].Len; len > 0 {
			data, err := readBytes(r, int(len))
			if err != nil {
				return err
			}
			m.Samples[i].Data = data
		}
	}

	return nil
}

func (m *Mod) Write(w io.Writer) error {
	if err := writeName(w, m.Name, LEN_NAME); err != nil {
		return err
	}

	// TODO write correct number of samples based on initials
	for i := 0; i < NUM_SAMPLES; i++ {
		if err := writeSampleHead(w, m.Samples[i]); err != nil {
			return err
		}
	}

	// TODO check range [1,128]
	if err := writeByte(w, m.Len); err != nil {
		return err
	}
	// TODO check =127
	if err := writeByte(w, 127); err != nil {
		return err
	}

	// TODO check values [0,63]
	if err := writeBytes(w, m.Sequence); err != nil {
		return err
	}

	mk := []byte("M.K.")
	if err := writeBytes(w, mk); err != nil {
		return err
	}

	for _, p := range m.Patterns {
		if err := writePattern(w, p); err != nil {
			return err
		}
	}

	for i := 0; i < NUM_SAMPLES; i++ {
		if err := writeBytes(w, m.Samples[i].Data); err != nil {
			return err
		}
	}

	return nil
}

func readSampleHead(r io.Reader) (Sample, error) {
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

func writeSampleHead(w io.Writer, s Sample) error {
	if err := writeName(w, s.Name, LEN_SAMPLE_NAME); err != nil {
		return err
	}
	if err := writeWyde(w, s.Len/2); err != nil {
		return err
	}
	if err := writeNibble(w, s.Tune); err != nil {
		return err
	}
	if err := writeByte(w, s.Volume); err != nil {
		return err
	}
	if err := writeWyde(w, s.RepeatStart/2); err != nil {
		return err
	}
	if err := writeWyde(w, s.RepeatLen/2); err != nil {
		return err
	}
	return nil
}

func readPattern(r io.Reader) (Pattern, error) {
	data, err := readBytes(r, 1024)
	if err != nil {
		return Pattern{}, err
	}
	return Pattern{data}, nil
}

func writePattern(w io.Writer, p Pattern) error {
	return writeBytes(w, p.Data)
}

func readName(r io.Reader, n int) ([]byte, error) {
	name, err := readBytes(r, n)
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

func writeName(w io.Writer, p []byte, n int) error {
	for len(p) < n {
		p = append(p, 0)
	}
	return writeBytes(w, p)
}

func readWyde(r io.Reader) (uint16, error) {
	b, err := readBytes(r, 2)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(b), nil
}

func writeWyde(w io.Writer, v uint16) error {
	p := make([]byte, 2)
	binary.BigEndian.PutUint16(p, v)
	return writeBytes(w, p)
}

func readNibble(r io.Reader) (int8, error) {
	b, err := readBytes(r, 1)
	if err != nil {
		return 0, err
	}
	return int8(b[0] & 0x0F), nil
}

func writeNibble(w io.Writer, v int8) error {
	return writeByte(w, byte(v))
}

func readByte(r io.Reader) (byte, error) {
	b, err := readBytes(r, 1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

func writeByte(w io.Writer, v byte) error {
	return writeBytes(w, []byte{v})
}

func readBytes(r io.Reader, len int) ([]byte, error) {
	b := make([]byte, len)
	_, err := io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func writeBytes(w io.Writer, p []byte) error {
	n, err := w.Write(p)
	if err != nil {
		return err
	}
	if n != len(p) {
		return errors.New("bad write")
	}
	return nil
}
