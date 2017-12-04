package midi

import (
	"fmt"
	"io/ioutil"
	"log"
)

type Parser struct {
	data              []byte
	position          int
	previousEventType uint8
	logger            *log.Logger
}

func (p *Parser) debugf(format string, v ...interface{}) {
	format = fmt.Sprintf("midi: [%v] %v", p.position, format)
	p.logger.Printf(format, v...)
}

func (p *Parser) debugln(v ...interface{}) {
	a := make([]interface{}, len(v)+1)
	a[0] = fmt.Sprintf("midi: [%v]", p.position)

	for i := 0; i < len(v); i++ {
		a[i+1] = v[i]
	}

	p.logger.Println(a...)
}

// Parse parses standard MIDI (*.mid) data.
func (p *Parser) Parse(stream []byte) (*MIDI, error) {
	p.logger.Printf("start parsing %v bytes\n", len(stream))

	formatType, numberOfTracks, timeDivision, err := p.parseHeader()
	if err != nil {
		return nil, err
	}

	tracks, err := p.parseTracks(numberOfTracks)
	if err != nil {
		return nil, err
	}

	midi := &MIDI{
		formatType:   formatType,
		timeDivision: &TimeDivision{value: timeDivision},
		Tracks:       tracks,
	}

	p.logger.Println("successfully done")

	return midi, nil
}

// parseHeader parses stream begins with MThd.
func (p *Parser) parseHeader() (formatType, numberOfTracks, timeDivision uint16, err error) {
	p.debugf("start parsing MThd")

	mthd := string(p.data[p.position:4])
	if mthd != "MThd" {
		return formatType, numberOfTracks, timeDivision, fmt.Errorf("midi: invalid chunk ID %v", mthd)
	}

	p.position += 4
	p.debugln("parsing MThd completed")

	p.debugln("start parsing header size")

	headerSize := p.data[p.position+3]
	if headerSize != 6 {
		return formatType, numberOfTracks, timeDivision, fmt.Errorf("midi: header size must be always 6 bytes (%v)", headerSize)
	}

	p.position += 4
	p.debugln("parsing header size completed")

	p.debugln("start parsing format type")

	formatType = uint16(p.data[p.position+1])
	if formatType > 3 {
		return formatType, numberOfTracks, timeDivision, fmt.Errorf("midi: format type should be 1, 2 or 3")
	}

	p.position += 2
	p.debugf("parsing format type completed (formatType=%v)", formatType)

	p.debugln("start parsing number of tracks")

	numberOfTracks = uint16(p.data[p.position])
	numberOfTracks = numberOfTracks << 8
	numberOfTracks += uint16(p.data[p.position+1])

	p.position += 2
	p.debugf("parsing number of tracks completed (%v)", numberOfTracks)

	p.debugln("start parsing time division")

	timeDivision = uint16(p.data[p.position])
	timeDivision = timeDivision << 8
	timeDivision += uint16(p.data[p.position+1])

	p.position += 2
	p.debugf("parsing time division completed (timeDivision = %v)", timeDivision)

	return formatType, numberOfTracks, timeDivision, nil
}

// parseTracks parses stream begins with MTrk.
func (p *Parser) parseTracks(numberOfTracks uint16) ([]*Track, error) {
	tracks := make([]*Track, numberOfTracks)

	for n := 0; n < int(numberOfTracks); n++ {
		p.debugln("start parsing MTrk")
		mtrk := string(p.data[p.position : p.position+4])
		if mtrk != "MTrk" {
			return nil, fmt.Errorf("midi: invalid track ID %v", mtrk)
		}

		p.position += 4
		p.debugln("parsing MTrk completed")

		p.debugln("start parsing size of track")

		chunkSize := uint32(p.data[p.position])
		chunkSize = chunkSize << 8
		chunkSize += uint32(p.data[p.position+1])
		chunkSize = chunkSize << 8
		chunkSize += uint32(p.data[p.position+2])
		chunkSize = chunkSize << 8
		chunkSize += uint32(p.data[p.position+3])

		p.position += 4
		p.debugf("parsing size of track completed (chunkSize=%v)", chunkSize)

		track, err := p.parseTrack()
		if err != nil {
			return nil, err
		}

		tracks[n] = track
		p.position += int(chunkSize)
	}

	return tracks, nil
}

// parseTrack parses stream begins with delta time and ends with end of track event.
func (p *Parser) parseTrack() (*Track, error) {
	sizeOfStream := len(p.data)
	track := &Track{
		Events: []Event{},
	}
	for {
		if p.position >= sizeOfStream {
			break
		}

		event, sizeOfEvent, err := p.parseEvent()
		if err != nil {
			return nil, err
		}
		track.Events = append(track.Events, event)
		p.position += sizeOfEvent

		switch event.(type) {
		case *EndOfTrackEvent:
			return track, nil
		}
	}

	return nil, fmt.Errorf("midi: missing end of track event")
}

// parseEvent parses stream begins with delta time.
func (p *Parser) parseEvent() (event Event, sizeOfEvent int, err error) {
	p.debugln("start parsing event")

	deltaTime, err := p.parseDeltaTime()
	if err != nil {
		return nil, 0, err
	}

	sizeOfDeltaTime := len(deltaTime.Quantity().value)
	eventType := p.data[sizeOfDeltaTime]

	if eventType < 0x80 && p.previousEventType >= 0x80 {
		eventType = p.previousEventType
	}
	switch eventType {
	case Meta:
		return p.parseMetaEvent(deltaTime)
	case SystemExclusive, DividedSystemExclusive:
		return p.parseSystemExclusiveEvent(deltaTime)
	default:
		return p.parseMIDIControlEvent(deltaTime, eventType)
	}
}

// parseMetaEvent parses stream begins with 0xff.
func (p *Parser) parseMetaEvent(deltaTime *DeltaTime) (event Event, sizeOfEvent int, err error) {
	q, err := p.parseQuantity()
	if err != nil {
		return nil, 0, err
	}

	offset := 2 + len(q.value)
	sizeOfData := int(q.Uint32())
	sizeOfEvent = len(deltaTime.Quantity().Value()) + offset + sizeOfData

	metaEventType := p.data[1]
	metaEventData := p.data[offset : offset+sizeOfData]

	switch metaEventType {
	case Text:
		event = &TextEvent{
			deltaTime: deltaTime,
			text:      metaEventData,
		}
	case CopyrightNotice:
		event = &CopyrightNoticeEvent{
			deltaTime: deltaTime,
			text:      metaEventData,
		}
	case SequenceOrTrackName:
		event = &SequenceOrTrackNameEvent{
			deltaTime: deltaTime,
			text:      metaEventData,
		}
	case InstrumentName:
		event = &InstrumentNameEvent{
			deltaTime: deltaTime,
			text:      metaEventData,
		}
	case Lyrics:
		event = &LyricsEvent{
			deltaTime: deltaTime,
			text:      metaEventData,
		}
	case Marker:
		event = &MarkerEvent{
			deltaTime: deltaTime,
			text:      metaEventData,
		}
	case CuePoint:
		event = &CuePointEvent{
			deltaTime: deltaTime,
			text:      metaEventData,
		}
	case MIDIPortPrefix:
		event = &MIDIPortPrefixEvent{
			deltaTime: deltaTime,
			port:      uint8(metaEventData[0]),
		}
	case MIDIChannelPrefix:
		event = &MIDIChannelPrefixEvent{
			deltaTime: deltaTime,
			channel:   uint8(metaEventData[0]),
		}
	case SetTempo:
		tempo := uint32(metaEventData[0])
		tempo = (tempo << 8) + uint32(metaEventData[1])
		tempo = (tempo << 8) + uint32(metaEventData[2])
		event = &SetTempoEvent{
			deltaTime: deltaTime,
			tempo:     tempo,
		}
	case SMPTEOffset:
		event = &SMPTEOffsetEvent{
			deltaTime: deltaTime,
			hour:      metaEventData[0],
			minute:    metaEventData[1],
			second:    metaEventData[2],
			frame:     metaEventData[3],
			subFrame:  metaEventData[4],
		}
	case TimeSignature:
		event = &TimeSignatureEvent{
			deltaTime:      deltaTime,
			numerator:      uint8(metaEventData[0]),
			denominator:    uint8(metaEventData[1]),
			metronomePulse: uint8(metaEventData[2]),
			quarterNote:    uint8(metaEventData[3]),
		}
	case KeySignature:
		event = &KeySignatureEvent{
			deltaTime: deltaTime,
			key:       int8(metaEventData[0]),
			scale:     uint8(metaEventData[1]),
		}
	case SequencerSpecific:
		event = &SequencerSpecificEvent{
			deltaTime: deltaTime,
			data:      metaEventData,
		}
	case EndOfTrack:
		event = &EndOfTrackEvent{
			deltaTime: deltaTime,
		}
	default:
		event = &AlienEvent{
			deltaTime:     deltaTime,
			metaEventType: metaEventType,
			data:          metaEventData,
		}
	}

	p.previousEventType = Meta
	p.debugf("parsing event completed (event = %v)", event)

	return event, sizeOfEvent, nil
}

// parseSystemExclusiveEvent parses stream begins with 0xf0 or 0xf7.
func (p *Parser) parseSystemExclusiveEvent(deltaTime *DeltaTime) (event Event, sizeOfEvent int, err error) {
	q, err := p.parseQuantity()
	if err != nil {
		return nil, 0, err
	}

	offset := 1 + len(q.value)
	sizeOfData := int(q.Uint32())
	sizeOfEvent = len(deltaTime.Quantity().value) + offset + sizeOfData
	eventType := p.data[0]

	switch eventType {
	case SystemExclusive:
		event = &SystemExclusiveEvent{
			deltaTime: deltaTime,
			data:      p.data[offset : offset+sizeOfData],
		}
	case DividedSystemExclusive:
		event = &DividedSystemExclusiveEvent{
			deltaTime: deltaTime,
			data:      p.data[offset : offset+sizeOfData],
		}
	}

	p.previousEventType = eventType
	p.debugf("parsing event completed (event = %v)", event)

	return event, sizeOfEvent, nil
}

// parseMIDIControlEvent parses stream begins with 0x8_...0xe_.
func (p *Parser) parseMIDIControlEvent(deltaTime *DeltaTime, eventType byte) (event Event, sizeOfEvent int, err error) {
	parameter := p.data[1:3]
	channel := uint8(eventType) & 0x0f
	sizeOfMIDIControlEvent := 3

	switch eventType & 0xf0 {
	case NoteOff:
		event = &NoteOffEvent{
			deltaTime: deltaTime,
			channel:   channel,
			note:      Note(parameter[0]),
			velocity:  parameter[1],
		}
	case NoteOn:
		event = &NoteOnEvent{
			deltaTime: deltaTime,
			channel:   channel,
			note:      Note(parameter[0]),
			velocity:  parameter[1],
		}
	case NoteAfterTouch:
		event = &NoteAfterTouchEvent{
			deltaTime: deltaTime,
			channel:   channel,
			note:      Note(parameter[0]),
			velocity:  uint8(parameter[1]),
		}
	case Controller:
		event = &ControllerEvent{
			deltaTime: deltaTime,
			channel:   channel,
			control:   Control(parameter[0]),
			value:     uint8(parameter[1]),
		}
	case ProgramChange:
		sizeOfMIDIControlEvent = 2
		event = &ProgramChangeEvent{
			deltaTime: deltaTime,
			channel:   channel,
			program:   GM(parameter[0]),
		}
	case ChannelAfterTouch:
		sizeOfMIDIControlEvent = 2
		event = &ChannelAfterTouchEvent{
			deltaTime: deltaTime,
			channel:   channel,
			velocity:  uint8(parameter[0]),
		}
	case PitchBend:
		pitch := uint16(parameter[0]&0x7f) << 7
		pitch += uint16(parameter[1] & 0x7f)
		event = &PitchBendEvent{
			deltaTime: deltaTime,
			channel:   channel,
			pitch:     pitch,
		}
	default:
		sizeOfMIDIControlEvent = 2
		event = &ContinuousControllerEvent{
			deltaTime: deltaTime,
			control:   uint8(p.data[0]),
			value:     uint8(p.data[1]),
		}
	}

	sizeOfEvent = len(deltaTime.Quantity().Value()) + sizeOfMIDIControlEvent

	p.previousEventType = eventType
	p.debugf("parsing event completed (event = %v)", event)

	return event, sizeOfEvent, nil
}

func (p *Parser) parseDeltaTime() (*DeltaTime, error) {
	q, err := p.parseQuantity()
	if err != nil {
		return nil, err
	}

	deltaTime := &DeltaTime{q}

	return deltaTime, nil
}

func (p *Parser) parseQuantity() (*Quantity, error) {
	if len(p.data) == 0 {
		return nil, fmt.Errorf("midi: stream is empty")
	}

	var i int
	q := &Quantity{}

	for {
		if i > 3 {
			return nil, fmt.Errorf("midi: maximum size of variable quantity is 4 bytes")
		}
		if len(p.data) < (i + 1) {
			return nil, fmt.Errorf("midi: missing next byte")
		}
		if p.data[i] < 0x80 {
			break
		}
		i++
	}

	q.value = make([]byte, i+1)
	copy(q.value, p.data)

	return q, nil
}

// SetLogger sets logger.
func (p *Parser) SetLogger(logger *log.Logger) *Parser {
	if logger != nil {
		p.logger = logger
	}

	return p
}

// NewParser returns Parser.
func NewParser(data []byte) *Parser {
	return &Parser{
		data:   data,
		logger: log.New(ioutil.Discard, "discard logging messages", 0),
	}
}
