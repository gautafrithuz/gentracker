package gentracker

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
	LEN_PATTERN     = 64
	NUM_CHANNELS    = 4
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

type Note struct {
	Sample uint8
	Period uint16
	Effect uint16
}

type Pattern [LEN_PATTERN][NUM_CHANNELS]Note

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
	m.Len = len

	_, err = readByte(r)
	if err != nil {
		return err
	}

	seq, err := readBytes(r, LEN_SEQUENCE)
	if err != nil {
		return err
	}
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

	if err := m.validate(); err != nil {
		return err
	}
	return nil
}

func (m *Mod) Write(w io.Writer) error {
	if err := m.validate(); err != nil {
		return err
	}

	if err := writeName(w, m.Name, LEN_NAME); err != nil {
		return err
	}

	for i := 0; i < NUM_SAMPLES; i++ {
		if err := writeSampleHead(w, m.Samples[i]); err != nil {
			return err
		}
	}

	if err := writeByte(w, m.Len); err != nil {
		return err
	}
	if err := writeByte(w, 127); err != nil {
		return err
	}

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
	var p Pattern
	for i := 0; i < LEN_PATTERN; i++ {
		for j := 0; j < NUM_CHANNELS; j++ {
			data, err := readBytes(r, 4)
			if err != nil {
				return p, err
			}
			p[i][j] = readNote(data)
		}
	}
	return p, nil
}

func writePattern(w io.Writer, p Pattern) error {
	for i := 0; i < LEN_PATTERN; i++ {
		for j := 0; j < NUM_CHANNELS; j++ {
			if err := writeBytes(w, writeNote(p[i][j])); err != nil {
				return err
			}
		}
	}
	return nil
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

func readNote(d []byte) Note {
	return Note{
		Sample: (d[0] & 0xF0) | ((d[2] & 0xF0) >> 4),
		Period: (uint16(d[0]&0x0F) << 8) | uint16(d[1]),
		Effect: (uint16(d[2]&0x0F) << 8) | uint16(d[3]),
	}
}

func writeNote(n Note) []byte {
	d := make([]byte, 4)
	d[0] = (n.Sample & 0xF0) | byte((n.Period>>8)&0x0F)
	d[1] = byte(n.Period & 0xFF)
	d[2] = ((n.Sample & 0x0F) << 4) | byte((n.Effect>>8)&0x0F)
	d[3] = byte(n.Effect & 0xFF)
	return d
}

func (m *Mod) validate() error {
	err := errors.New("invalid mod")
	if m.Len < 1 || m.Len > 128 {
		return err
	}
	for _, v := range m.Sequence {
		if v < 0 || v > 63 {
			return err
		}
	}
	for _, sample := range m.Samples {
		if sample.Volume > 64 {
			return err
		}
	}
	return nil
}
