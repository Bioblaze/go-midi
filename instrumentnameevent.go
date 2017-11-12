package midi

import "fmt"

// InstrumentNameEvent corresponds to instrument name event.
type InstrumentNameEvent struct {
	deltaTime *DeltaTime
	text      []byte
}

// DeltaTime returns delta time of instrument name event.
func (e *InstrumentNameEvent) DeltaTime() *DeltaTime {
	if e.deltaTime == nil {
		e.deltaTime = &DeltaTime{}
	}
	return e.deltaTime
}

// String returns string representation of instrument name event.
func (e *InstrumentNameEvent) String() string {
	return fmt.Sprintf("&InstrumentNameEvent{text: \"%v\"}", string(e.Text()))
}

// Serialize serializes instrument name event.
func (e *InstrumentNameEvent) Serialize() []byte {
	bs := []byte{}
	bs = append(bs, e.DeltaTime().Quantity().Value()...)
	bs = append(bs, Meta, InstrumentName)

	q := &Quantity{}
	q.SetUint32(uint32(len(e.Text())))
	bs = append(bs, q.Value()...)
	bs = append(bs, e.Text()...)

	return bs
}

// SetText sets text.
func (e *InstrumentNameEvent) SetText(text []byte) error {
	if len(text) > 0xfffffff {
		return fmt.Errorf("midi: maximum size of text is 256 MB")
	}
	e.text = text

	return nil
}

// Text returns text.
func (e *InstrumentNameEvent) Text() []byte {
	if e.text == nil {
		e.text = []byte{}
	}

	text := make([]byte, len(e.text))
	copy(text, e.text)

	return text
}

// NewInstrumentNameEvent returns InstrumentNameEvent with the given parameter.
func NewInstrumentNameEvent(deltaTime *DeltaTime, text []byte) (*InstrumentNameEvent, error) {
	var err error

	event := &InstrumentNameEvent{}
	event.deltaTime = deltaTime

	err = event.SetText(text)
	if err != nil {
		return nil, err
	}
	return event, nil
}
