package music

import (
	"log"
	"os"

	beep "github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/flac"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/vorbis"
	"github.com/gopxl/beep/v2/wav"
)

// getStreamer is responsible for decoding and returning the appropriate streamer and format based on file type
func GetStreamer(music MusicMetaData) (beep.StreamSeekCloser, beep.Format) {
	osfile, err := os.Open(music.filepath)
	if err != nil {
		log.Fatal("Error opening file: ", err)
	}

	var streamer beep.StreamSeekCloser
	var format beep.Format

	switch music.fileformat {
	case "audio/mp3":
		streamer, format, err = mp3.Decode(osfile)
	case "audio/wav":
		streamer, format, err = wav.Decode(osfile)
	case "audio/flac":
		streamer, format, err = flac.Decode(osfile)
	case "audio/ogg":
		streamer, format, err = vorbis.Decode(osfile)
	default:
		log.Fatal("Unsupported file type")
	}

	if err != nil {
		log.Fatal("Error while decoding file: ", err)
	}

	return streamer, format
}

type MusicState struct {
	Audiofile  *os.File
	SampleRate beep.SampleRate
	Streamer   beep.StreamSeekCloser
	format     beep.Format
	ctrl       *beep.Ctrl
	resampler  *beep.Resampler
	volume     *effects.Volume
}

// NewMState initializes a new musicState object
func NewMState(filepath string, format beep.Format, streamer beep.StreamSeekCloser) *MusicState {
	osaudiofile, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	ctrl := &beep.Ctrl{Streamer: beep.Loop(1, streamer)}
	resampler := beep.ResampleRatio(4, 1, ctrl)
	volume := &effects.Volume{Streamer: resampler, Base: 2}
	return &MusicState{osaudiofile, format.SampleRate, streamer, format, ctrl, resampler, volume}
}

// play starts playing the audio
func (ms *MusicState) Play(music MusicMetaData) {
	ms.Streamer, ms.format = GetStreamer(music)
	done := make(chan bool)
	speaker.Play(beep.Seq(ms.volume, beep.Callback(func() {
		done <- true
	})))
	<-done
}

func (ms *MusicState) SetVolume(delta float64) {
	speaker.Lock()
	ms.volume.Volume += delta
	speaker.Unlock()
}
func (ms *MusicState) PauseorUnpause() {
	speaker.Lock()
	ms.ctrl.Paused = !ms.ctrl.Paused
	speaker.Unlock()
}

func (ms *MusicState) Stop() {
	if ms.ctrl != nil {
		ms.ctrl.Streamer = nil
		ms.ctrl = nil
	}
if ms.Audiofile != nil {
		_ = ms.Audiofile.Close()
		ms.Audiofile = nil
	}
	if ms.Streamer != nil {
		_ = ms.Streamer.Close()
		ms.Streamer = nil
	}
	if ms.resampler != nil {
		ms.resampler = nil
	}
	if ms.volume != nil {
		ms.volume = nil
	}
}
