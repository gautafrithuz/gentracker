package gentracker

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestRead(t *testing.T) {
	var m Mod

	f, err := os.Open("example.mod")
	if err != nil {
		panic(err)
	}
	err = m.Read(f)
	if err != nil {
		t.Error(err)
	}

	if string(m.Name) != "01-WARM" {
		t.Errorf("Name: %s", m.Name)
	}

	if len(m.Samples) != NUM_SAMPLES {
		t.Errorf("Num Samples: %d", len(m.Samples))
	}
	if string(m.Samples[8].Name) != "TECH-CRASH" {
		t.Errorf("Sample Name: %s", m.Samples[8].Name)
	}
	if m.Samples[8].Len != 15458 {
		t.Errorf("Sample Len: %d", m.Samples[8].Len)
	}
	if m.Samples[8].Tune != 0 {
		t.Errorf("Sample Tune: %d", m.Samples[8].Tune)
	}
	if m.Samples[8].Volume != 48 {
		t.Errorf("Sample Volume: %d!=%d", m.Samples[8].Volume)
	}
	if m.Samples[8].RepeatStart != 13534 {
		t.Errorf("Sample Repeat Start: %d", m.Samples[8].RepeatStart)
	}
	if m.Samples[8].RepeatLen != 1924 {
		t.Errorf("Sample Repeat Start: %d", m.Samples[8].RepeatLen)
	}
	lens := []uint16{52176, 25764, 27884, 40070, 24560, 30800, 256, 20722, 15458, 18724, 17640, 9014, 8820, 18724, 30646}
	for len(lens) < NUM_SAMPLES {
		lens = append(lens, 0)
	}
	for i := 0; i < NUM_SAMPLES; i++ {
		if m.Samples[i].Len != lens[i] {
			t.Errorf("Sample Len %d: %d", i, m.Samples[i].Len)
		}
	}

	if m.Len != 48 {
		t.Errorf("Len: %d", m.Len)
	}

	seq := []byte{0x13, 0x03, 0x02, 0x03, 0x04, 0x05, 0x04, 0x06, 0x07, 0x08, 0x09, 0x08, 0x0C, 0x08, 0x0C, 0x08, 0x0A, 0x0B, 0x0A, 0x0B, 0x00, 0x01, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x11, 0x12, 0x00, 0x01, 0x16, 0x14, 0x15, 0x17, 0x18, 0x17, 0x19, 0x19, 0x19, 0x19, 0x1C, 0x1D, 0x1E, 0x1D, 0x1B, 0x1A}
	for len(seq) < LEN_SEQUENCE {
		seq = append(seq, 0)
	}
	if !reflect.DeepEqual(m.Sequence, seq) {
		t.Errorf("Mod Seq")
	}
}

func TestWrite(t *testing.T) {
	var m Mod

	f, err := os.Open("example.mod")
	if err != nil {
		panic(err)
	}
	var expected bytes.Buffer
	tee := io.TeeReader(f, &expected)
	err = m.Read(tee)
	if err != nil {
		t.Error(err)
	}

	var actual bytes.Buffer
	err = m.Write(&actual)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(actual.Bytes(), expected.Bytes()) {
		t.Error("Write not binary equal!")
	}
}

func TestNote(t *testing.T) {
	expected := []byte{0x0F, 0x96, 0xA5, 0xC3}
	n := readNote(expected)
	if n.Sample != 0x0A {
		t.Errorf("Incorrect Note Sample")
	}
	if n.Period != 0xF96 {
		t.Errorf("Incorrect Note Period")
	}
	if n.Effect != 0x5C3 {
		t.Errorf("Incorrect Note Effect")
	}

	actual := writeNote(n)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Note read/write")
	}
}
