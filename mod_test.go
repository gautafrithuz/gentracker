package amiga

import (
	"os"
	"reflect"
	"testing"
)

func TestRead(t *testing.T) {
	f, err := os.Open("example.mod")
	if err != nil {
		panic(err)
	}
	mod, err := Read(f)
	if err != nil {
		t.Error(err)
	}

	if string(mod.Name) != "01-WARM" {
		t.Errorf("Name: %s", mod.Name)
	}

	if len(mod.Samples) != NUM_SAMPLES {
		t.Errorf("Num Samples: %d", len(mod.Samples))
	}
	if string(mod.Samples[8].Name) != "TECH-CRASH" {
		t.Errorf("Sample Name: %s", mod.Samples[8].Name)
	}
	if mod.Samples[8].Len != 15458 {
		t.Errorf("Sample Len: %d", mod.Samples[8].Len)
	}
	if mod.Samples[8].Tune != 0 {
		t.Errorf("Sample Tune: %d", mod.Samples[8].Tune)
	}
	if mod.Samples[8].Volume != 48 {
		t.Errorf("Sample Volume: %d!=%d", mod.Samples[8].Volume)
	}
	if mod.Samples[8].RepeatStart != 13534 {
		t.Errorf("Sample Repeat Start: %d", mod.Samples[8].RepeatStart)
	}
	if mod.Samples[8].RepeatLen != 1924 {
		t.Errorf("Sample Repeat Start: %d", mod.Samples[8].RepeatLen)
	}
	lens := []uint16{52176, 25764, 27884, 40070, 24560, 30800, 256, 20722, 15458, 18724, 17640, 9014, 8820, 18724, 30646}
	for len(lens) < NUM_SAMPLES {
		lens = append(lens, 0)
	}
	for i := 0; i < NUM_SAMPLES; i++ {
		if mod.Samples[i].Len != lens[i] {
			t.Errorf("Sample Len %d: %d", i, mod.Samples[i].Len)
		}
	}

	if mod.Len != 48 {
		t.Errorf("Len: %d", mod.Len)
	}

	seq := []byte{0x13, 0x03, 0x02, 0x03, 0x04, 0x05, 0x04, 0x06, 0x07, 0x08, 0x09, 0x08, 0x0C, 0x08, 0x0C, 0x08, 0x0A, 0x0B, 0x0A, 0x0B, 0x00, 0x01, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x11, 0x12, 0x00, 0x01, 0x16, 0x14, 0x15, 0x17, 0x18, 0x17, 0x19, 0x19, 0x19, 0x19, 0x1C, 0x1D, 0x1E, 0x1D, 0x1B, 0x1A}
	for len(seq) < LEN_SEQUENCE {
		seq = append(seq, 0)
	}
	if !reflect.DeepEqual(mod.Sequence, seq) {
		t.Errorf("Mod Seq")
	}
}
