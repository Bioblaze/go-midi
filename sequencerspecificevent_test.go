package midi

import "testing"

func TestSequencerSpecificEvent_DeltaTime(t *testing.T) {
	event := &SequencerSpecificEvent{}
	dt := event.DeltaTime()
	if dt == nil {
		t.Fatal("DeltaTime() don't return nil")
	}
}

func TestSequencerSpecificEvent_String(t *testing.T) {
	event, err := NewSequencerSpecificEvent(nil, []byte{0x12, 0x34, 0x56})
	if err != nil {
		t.Fatal(err)
	}

	expected := "&SequencerSpecificEvent{data: 3 bytes}"
	actual := event.String()
	if expected != actual {
		t.Fatalf("expected: %v actual: %v", expected, actual)
	}
}

func TestSequencerSpecificEvent_Serialize(t *testing.T) {
	event, err := NewSequencerSpecificEvent(nil, []byte{0x12, 0x34, 0x56})
	if err != nil {
		t.Fatal(err)
	}

	expected := []byte{0x00, 0xff, 0x7f, 0x03, 0x12, 0x34, 0x56}
	actual := event.Serialize()

	if len(expected) != len(actual) {
		t.Fatalf("expected: %v bytes actual: %v bytes", len(expected), len(actual))
	}
	for i, e := range expected {
		a := actual[i]
		if e != a {
			t.Fatalf("expected[%v] = 0x%x actual[%v] = 0x%x", i, e, i, a)
		}
	}
}

func TestSequencerSpecificEvent_SetData(t *testing.T) {
	event := &SequencerSpecificEvent{}
	data := make([]byte, 0x10000000)

	err := event.SetData(data)
	if err == nil {
		t.Fatalf("err must not be nil")
	}

	data = make([]byte, 0xfffffff)
	err = event.SetData(data)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSequencerSpecificEvent_Data(t *testing.T) {
	event := &SequencerSpecificEvent{data: []byte{0x12, 0x34, 0x56}}

	expected := []byte{0x12, 0x34, 0x56}
	actual := event.Data()

	if len(expected) != len(actual) {
		t.Fatalf("expected: %v bytes actual: %v bytes", len(expected), len(actual))
	}
	for i, e := range expected {
		a := actual[i]
		if e != a {
			t.Fatalf("expected[%v] = 0x%x actual[%v] = 0x%x", i, e, i, a)
		}
	}
}

func TestNewSequencerSpecificEvent(t *testing.T) {
	event, err := NewSequencerSpecificEvent(nil, []byte{0x12, 0x34, 0x56})
	if err != nil {
		t.Fatal(err)
	}
	expected := []byte{0x12, 0x34, 0x56}
	actual := event.data

	if len(expected) != len(actual) {
		t.Fatalf("expected: %v bytes actual: %v bytes", len(expected), len(actual))
	}
	for i, e := range expected {
		a := actual[i]
		if e != a {
			t.Fatalf("expected[%v] = 0x%x actual[%v] = 0x%x", i, e, i, a)
		}
	}
}