package segmago

const (
	amountToStore    = 1 * 4 // must be power of 2
	samplesPerSecond = 44100

	ntscClocksPerSecond = 3579545
)

type sn76489 struct {
	buffer apuCircleBuf

	Sounds         [4]sound
	LatchedSound   *sound
	LatchIsForData bool

	ClocksPerSample int

	SampleSum      int
	SampleSumCount int

	lastOutput          float32
	lastCorrectedOutput float32

	Clock int
}

const apuCircleBufSize = amountToStore

// NOTE: size must be power of 2
type apuCircleBuf struct {
	writeIndex uint
	readIndex  uint
	buf        [apuCircleBufSize]byte
}

func (c *apuCircleBuf) write(bytes []byte) (writeCount int) {
	for _, b := range bytes {
		if c.full() {
			return writeCount
		}
		c.buf[c.mask(c.writeIndex)] = b
		c.writeIndex++
		writeCount++
	}
	return writeCount
}
func (c *apuCircleBuf) read(preSizedBuf []byte) []byte {
	readCount := 0
	for i := range preSizedBuf {
		if c.size() == 0 {
			break
		}
		preSizedBuf[i] = c.buf[c.mask(c.readIndex)]
		c.readIndex++
		readCount++
	}
	return preSizedBuf[:readCount]
}
func (c *apuCircleBuf) mask(i uint) uint { return i & (uint(len(c.buf)) - 1) }
func (c *apuCircleBuf) size() uint       { return c.writeIndex - c.readIndex }
func (c *apuCircleBuf) full() bool       { return c.size() == uint(len(c.buf)) }

type sound struct {
	Volume  byte
	Data    uint16
	Counter uint16
	Output  byte

	IsNoise bool
}

func (s *sn76489) init() {
	for i := range s.Sounds {
		s.Sounds[i].Volume = 0x0f
	}
	s.Sounds[3].IsNoise = true
	s.LatchedSound = &s.Sounds[0]

	f := ntscClocksPerSecond / samplesPerSecond
	s.ClocksPerSample = int(f)
}

func (s *sn76489) runCycle() {

	if !s.buffer.full() {

		if s.Clock == 0 {
			for i := range s.Sounds {
				s.Sounds[i].runCycle()
			}
		}
		s.Clock = (s.Clock + 1) & 0x0f

		sum := 0
		for i := range s.Sounds {
			sound := &s.Sounds[i]
			sum += int(sound.Output * (15 - sound.Volume))
		}

		s.SampleSum += sum
		s.SampleSumCount++
		if s.SampleSumCount >= s.ClocksPerSample {

			sum := float32(s.SampleSum) / 60.0 // 4 channels, 15 vol levels

			output := sum / float32(s.SampleSumCount)

			s.SampleSum = 0
			s.SampleSumCount = 0

			// dc blocker to center waveform
			correctedOutput := output - s.lastOutput + 0.995*s.lastCorrectedOutput
			s.lastCorrectedOutput = correctedOutput
			s.lastOutput = output
			output = correctedOutput

			sample := int16(output * 32767.0)
			sampleLo := byte(sample & 0xff)
			sampleHi := byte(sample >> 8)
			s.buffer.write([]byte{
				sampleLo, sampleHi,
				sampleLo, sampleHi,
			})
		}
	}
}

func (s *sound) runCycle() {
	if s.IsNoise {
	} else {
		s.Counter--
		if s.Counter == 0 {
			s.Counter = s.Data
			s.Output ^= 1
		}
		if s.Data == 0 || s.Data == 1 {
			s.Output = 1
		}
	}
}

func (s *sound) setCounterFromData() {
	if s.IsNoise {
	} else {
		s.Counter = s.Data
	}
}

func (s *sn76489) sendByte(b byte) {
	if b&0x80 > 0 {
		i := b >> 5 & 3
		s.LatchedSound = &s.Sounds[i]
		s.LatchIsForData = b&0x10 == 0
		if s.LatchIsForData {
			s.Sounds[i].Data &^= 0x0f
			s.Sounds[i].Data |= uint16(b & 0x0f)
		} else {
			s.Sounds[i].Volume &^= 0x0f
			s.Sounds[i].Volume |= b & 0x0f
		}

	} else {
		latch := s.LatchedSound
		if s.LatchIsForData {
			if latch.IsNoise {
				latch.Data &^= 0x0f
				latch.Data |= uint16(b & 0x0f)
			} else {
				latch.Data &^= 0x3f0
				latch.Data |= uint16(b&0x3f) << 4
			}
		} else {
			latch.Volume &^= 0x0f
			latch.Volume |= b & 0x0f
		}
	}
}
