package audio_wav

// Metadata represents optional metadata added to the wav file.
type Metadata struct {
	SamplerInfo  *SamplerInfo
	Artist       string
	Comments     string
	Copyright    string
	CreationDate string
	Engineer     string
	Technician   string
	Genre        string
	Keywords     string
	Medium       string
	Title        string
	Product      string
	Subject      string
	Software     string
	Source       string
	Location     string
	TrackNbr     string
	CuePoints    []*CuePoint

	Lyrics     string // 歌词文本
	CoverImage []byte // 封面图像二进制数据
}

// SamplerInfo is extra metadata pertinent to a sampler type usage.
type SamplerInfo struct {
	Manufacturer      [4]byte
	Product           [4]byte
	SamplePeriod      uint32
	MIDIUnityNote     uint32
	MIDIPitchFraction uint32
	SMPTEFormat       uint32
	SMPTEOffset       uint32
	NumSampleLoops    uint32
	Loops             []*SampleLoop
}

// SampleLoop indicates a loop and its properties within the audio file
type SampleLoop struct {
	CuePointID [4]byte
	Type       uint32
	Start      uint32
	End        uint32
	Fraction   uint32
	PlayCount  uint32
}
