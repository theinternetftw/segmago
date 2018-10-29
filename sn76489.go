package segmago

const (
	amountToStore    = 16 * 512 * 4 // must be power of 2
	samplesPerSecond = 44100

	ntscClocksPerSecond = 3579545
)

type sn76489 struct {
	buffer apuCircleBuf

	Sounds         [4]sound
	LatchedSound   *sound
	LatchIsForData bool

	ClocksPerSample int32

	SumLeft        int32
	SumRight       int32
	SampleSumCount int32

	lastOutputLeft           float32
	lastOutputRight          float32
	lastCorrectedOutputLeft  float32
	lastCorrectedOutputRight float32

	StereoMixerReg byte

	Clock int32
}

const apuCircleBufSize = amountToStore

// NOTE: size must be power of 2
type apuCircleBuf struct {
	writeIndex uint32
	readIndex  uint32
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
func (c *apuCircleBuf) mask(i uint32) uint32 { return i & (uint32(len(c.buf)) - 1) }
func (c *apuCircleBuf) size() uint32         { return c.writeIndex - c.readIndex }
func (c *apuCircleBuf) full() bool           { return c.size() == uint32(len(c.buf)) }

type sound struct {
	Volume  byte
	Data    uint16
	Counter uint16
	Output  byte

	IsNoise    bool
	NoiseClock bool
	LFSR       uint16
}

func (s *sn76489) init() {
	for i := range s.Sounds {
		s.Sounds[i].Volume = 0x0f
	}
	s.Sounds[3].IsNoise = true
	s.LatchedSound = &s.Sounds[0]

	f := ntscClocksPerSecond / samplesPerSecond
	s.ClocksPerSample = int32(f)

	s.StereoMixerReg = 0xff
}

func (s *sn76489) readSoundBuffer(toFill []byte) {
	if int(s.buffer.size()) < len(toFill) {
		//fmt.Println("audSize:", s.buffer.size(), "len(toFill)", len(toFill))
	}
	for int(s.buffer.size()) < len(toFill) {
		// stretch sound to fill buffer to avoid click
		s.genSample()
	}
	s.buffer.read(toFill)
}

func (s *sn76489) genSample() {
	if s.Clock == 0 {
		for i := range s.Sounds {
			s.runSoundCycle(&s.Sounds[i])
		}

		for i := range s.Sounds {
			sound := &s.Sounds[i]
			sample := int32(sound.Output * (15 - sound.Volume))
			if s.StereoMixerReg>>uint32(i)&1 > 0 {
				s.SumLeft += sample
			}
			if s.StereoMixerReg>>uint32(4+i)&1 > 0 {
				s.SumRight += sample
			}
		}

		s.SampleSumCount++
		if s.SampleSumCount >= s.ClocksPerSample>>4 {

			outLeft := float32(s.SumLeft)
			outLeft /= 15.0 * float32(s.SampleSumCount*4) // 15 vol levels

			outRight := float32(s.SumRight)
			outRight /= 15.0 * float32(s.SampleSumCount*4) // 15 vol levels

			s.SumLeft = 0
			s.SumRight = 0
			s.SampleSumCount = 0

			// dc blocker to center waveform
			correctedOutputLeft := outLeft - s.lastOutputLeft + 0.995*s.lastCorrectedOutputLeft
			s.lastCorrectedOutputLeft = correctedOutputLeft
			s.lastOutputLeft = outLeft
			outLeft = correctedOutputLeft

			// dc blocker to center waveform
			correctedOutputRight := outRight - s.lastOutputRight + 0.995*s.lastCorrectedOutputRight
			s.lastCorrectedOutputRight = correctedOutputRight
			s.lastOutputRight = outRight
			outRight = correctedOutputRight

			sampleLeft := int16(outLeft * 32767.0)
			sampleRight := int16(outRight * 32767.0)
			s.buffer.write([]byte{
				byte(sampleLeft & 0xff), byte(sampleLeft >> 8),
				byte(sampleRight & 0xff), byte(sampleRight >> 8),
			})
		}
	}
	s.Clock = (s.Clock + 1) & 0x0f
}

var newBufFull = false

func (s *sn76489) runCycle() {

	if !s.buffer.full() {
		s.genSample()
		newBufFull = true
	} else if newBufFull {
		//fmt.Println("sn buf full!")
		newBufFull = false
	}
}

func (s *sn76489) runSoundCycle(snd *sound) {
	if snd.IsNoise {
		snd.Counter--
		if snd.Counter == 0 {
			tbl := []uint16{
				0x10, 0x20, 0x40, s.Sounds[2].Data,
			}
			snd.Counter = tbl[snd.Data&3]
			snd.NoiseClock = !snd.NoiseClock
			if snd.NoiseClock {
				snd.Output = byte(snd.LFSR & 1)
				// TODO: LFSR variants (e.g. bbc micro has 15-bit LFSR, taps bits 0 and 1)
				var newBit uint16
				if snd.Data&0x04 > 0 {
					newBit = ((snd.LFSR & 1) ^ (snd.LFSR >> 3)) & 1
				} else {
					newBit = snd.LFSR & 1
				}
				snd.LFSR >>= 1
				snd.LFSR |= newBit << 15
			}
		}
	} else {
		snd.Counter--
		if snd.Counter == 0 {
			snd.Counter = snd.Data
			snd.Output ^= 1
		}
		if snd.Data == 0 || snd.Data == 1 {
			snd.Output = 1
		}
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
			s.Sounds[i].LFSR = 0x8000
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
				latch.LFSR = 0x8000
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
