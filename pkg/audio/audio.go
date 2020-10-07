package audio

import (
	"pokered/pkg/store"
	"pokered/pkg/util"

	"github.com/hajimehoshi/ebiten/audio"
)

const (
	sampleRate     = 44100
	stopSound  int = -1
)

var audioContext, _ = audio.NewContext(sampleRate)

var FadeOut = struct {
	Control uint
	Counter uint
	Reload  uint
}{
	Reload: 10,
}

// NewMusicID Music ID played on current music fadeout is completed
var NewMusicID int

// FadeOutAudio fadeout process called in every vBlank
func FadeOutAudio() {
	preVolume := Volume
	defer func() {
		if CurMusic != nil && CurMusic.IsPlaying() && preVolume != Volume {
			CurMusic.SetVolume(float64(Volume) / 7)
		}
	}()

	if FadeOut.Control == 0 {
		if util.ReadBit(store.D72C, 1) {
			return
		}
		setVolumeMax()
	}

	// fade out
	if FadeOut.Counter > 0 {
		FadeOut.Counter--
		return
	}

	// counterReachedZero
	{
		FadeOut.Counter = FadeOut.Reload

		// fadeOutComplete
		if Volume == 0 {
			// start next music
			FadeOut.Control = 0
			if CurMusic != nil {
				CurMusic.Close()
			}
			PlayMusic(NewMusicID)
			return
		}

		decrementVolume()
	}
}
