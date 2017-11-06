package midi

type DeltaTime struct {
	*Quantity
}

func (d *DeltaTime) Value() int {
	return int(d.value[0])
}

func (d *DeltaTime) SetRawValue(value []byte) {
	d.value = value
}

func (d *DeltaTime) serialize() []byte {
	return d.value
}

func parseDeltaTime(stream []byte) (*DeltaTime, error) {
	q, err := parseQuantity(stream)
	if err != nil {
		return nil, err
	}

	deltaTime := &DeltaTime{q}

	return deltaTime, nil
}

/*
func parseDeltaTime(stream []byte) (*DeltaTime, error) {
	if len(stream) == 0 {
		return nil, fmt.Errorf("midi.parseDeltaTime: stream is empty")
	}

	var i int
	dt := &DeltaTime{}

	for {
		if i > 3 {
			return nil, fmt.Errorf("midi.parseDeltaTime: maximum size of delta time is 4 bytes")
		}
		if len(stream) < (i + 1) {
			return nil, fmt.Errorf("midi.parseDeltaTime: missing next byte (stream=%+v)", stream)
		}
		if stream[i] < 0x80 {
			break
		}
		i++
	}

	dt.value = make([]byte, i+1)
	copy(dt.value, stream)

	return dt, nil
}
*/
