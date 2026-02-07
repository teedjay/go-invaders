package main

import (
	"bytes"
	"errors"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/gotracker/playback/format"
	"github.com/gotracker/playback/mixing"
	"github.com/gotracker/playback/mixing/sampling"
	"github.com/gotracker/playback/output"
	"github.com/gotracker/playback/player/feature"
	"github.com/gotracker/playback/player/machine"
	"github.com/gotracker/playback/player/machine/settings"
	"github.com/gotracker/playback/player/sampler"
	"github.com/gotracker/playback/song"
)

const (
	audioSampleRate = 44100
	sfxVolume       = 0.5
)

type SFX struct {
	Swoosh    []byte
	Wheee     []byte
	Explosion []byte
	UFOWooo   []byte
	SuperDrum []byte
}

func (g *Game) playSfx(data []byte, volume float64) {
	if g.AudioCtx == nil || len(data) == 0 {
		return
	}
	player := audio.NewPlayerFromBytes(g.AudioCtx, data)
	if player == nil {
		return
	}
	player.SetVolume(volume)
	player.Play()
}

func (g *Game) startMusic(path string) {
	if g.AudioCtx == nil || g.MusicPlayer != nil {
		return
	}
	data, err := renderMOD(path, audioSampleRate)
	if err != nil || len(data) == 0 {
		return
	}
	loop := audio.NewInfiniteLoop(bytes.NewReader(data), int64(len(data)))
	player, err := audio.NewPlayer(g.AudioCtx, loop)
	if err != nil {
		return
	}
	player.SetVolume(0.3)
	player.Play()
	g.MusicPlayer = player
}

func (g *Game) setMusicVolume(v float64) {
	if g.MusicPlayer == nil {
		return
	}
	g.MusicPlayer.SetVolume(v)
}

func (g *Game) stopMusic() {
	if g.MusicPlayer == nil {
		return
	}
	g.MusicPlayer.Close()
	g.MusicPlayer = nil
}

func (g *Game) startUFOLoop() {
	if g.UFOLoop != nil {
		return
	}
	if g.AudioCtx == nil || len(g.Sfx.UFOWooo) == 0 {
		return
	}
	loop := audio.NewInfiniteLoop(bytes.NewReader(g.Sfx.UFOWooo), int64(len(g.Sfx.UFOWooo)))
	player, err := audio.NewPlayer(g.AudioCtx, loop)
	if err != nil {
		return
	}
	player.SetVolume(sfxVolume * 0.35)
	player.Play()
	g.UFOLoop = player
}

func (g *Game) stopUFOLoop() {
	if g.UFOLoop == nil {
		return
	}
	g.UFOLoop.Close()
	g.UFOLoop = nil
}

func (g *Game) startSuperLoop() {
	if g.SuperLoop != nil {
		return
	}
	if g.AudioCtx == nil || len(g.Sfx.SuperDrum) == 0 {
		return
	}
	loop := audio.NewInfiniteLoop(bytes.NewReader(g.Sfx.SuperDrum), int64(len(g.Sfx.SuperDrum)))
	player, err := audio.NewPlayer(g.AudioCtx, loop)
	if err != nil {
		return
	}
	player.SetVolume(sfxVolume * 0.45)
	player.Play()
	g.SuperLoop = player
}

func (g *Game) stopSuperLoop() {
	if g.SuperLoop == nil {
		return
	}
	g.SuperLoop.Close()
	g.SuperLoop = nil
}

func (g *Game) playWheee() *audio.Player {
	if g.AudioCtx == nil || len(g.Sfx.Wheee) == 0 {
		return nil
	}
	player := audio.NewPlayerFromBytes(g.AudioCtx, g.Sfx.Wheee)
	if player == nil {
		return nil
	}
	player.SetVolume(sfxVolume * 0.6)
	player.Play()
	return player
}

func (g *Game) updateWheee(player *audio.Player, startY, y float64) {
	if player == nil {
		return
	}
	progress := 0.0
	if startY > 0 {
		progress = (startY - y) / startY
	}
	fade := 1.0 - progress*1.2
	if fade < 0 {
		fade = 0
	}
	player.SetVolume(sfxVolume * 0.6 * fade)
	if fade == 0 {
		g.stopWheee(player)
	}
}

func (g *Game) stopWheee(player *audio.Player) {
	if player == nil {
		return
	}
	player.Close()
}

func genSwoosh(sampleRate int, seconds float64) []byte {
	n := int(float64(sampleRate) * seconds)
	data := make([]byte, n*2)
	cutoff := 0.2
	prev := 0.0
	for i := 0; i < n; i++ {
		t := float64(i) / float64(n)
		env := 1.0 - t
		noise := (rand.Float64()*2 - 1)
		prev = prev + cutoff*(noise-prev)
		s := prev * env * 0.6
		writeSample(data, i, s)
	}
	return data
}

func genWheee(sampleRate int, seconds float64) []byte {
	n := int(float64(sampleRate) * seconds)
	data := make([]byte, n*2)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(sampleRate)
		freq := 440.0 + 120.0*math.Sin(t*2.0)
		s := math.Sin(2*math.Pi*freq*t) * 0.35
		writeSample(data, i, s)
	}
	return data
}

func genExplosion(sampleRate int, seconds float64) []byte {
	n := int(float64(sampleRate) * seconds)
	data := make([]byte, n*2)
	cutoff := 0.08
	prev := 0.0
	for i := 0; i < n; i++ {
		t := float64(i) / float64(n)
		env := math.Pow(1.0-t, 1.8)
		noise := (rand.Float64()*2 - 1)
		prev = prev + cutoff*(noise-prev)
		thump := math.Sin(2*math.Pi*60*float64(i)/float64(sampleRate)) * (1.0 - t)
		crack := (rand.Float64()*2 - 1) * math.Exp(-t*18)
		sparkle := math.Sin(2*math.Pi*1800*float64(i)/float64(sampleRate)) * math.Pow(t, 2) * 0.2
		s := (prev*0.9 + thump*1.2 + crack*0.7 + sparkle) * env
		writeSample(data, i, s)
	}
	return data
}

func genUFOWooo(sampleRate int, seconds float64) []byte {
	n := int(float64(sampleRate) * seconds)
	data := make([]byte, n*2)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(sampleRate)
		mod := 0.5 + 0.5*math.Sin(t*2.2)
		freq := 220.0 + 80.0*math.Sin(t*0.7)
		s := math.Sin(2*math.Pi*freq*t) * (0.4 + 0.3*mod)
		writeSample(data, i, s)
	}
	return data
}

func genSuperDrum(sampleRate int, seconds float64) []byte {
	n := int(float64(sampleRate) * seconds)
	data := make([]byte, n*2)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(sampleRate)
		beat := math.Mod(t, 0.25)
		env := math.Exp(-beat * 18)
		thump := math.Sin(2*math.Pi*110*t) * env
		click := (rand.Float64()*2 - 1) * math.Exp(-beat*60) * 0.2
		s := (thump*0.9 + click) * 0.9
		writeSample(data, i, s)
	}
	return data
}

func writeSample(buf []byte, i int, s float64) {
	if s > 1 {
		s = 1
	}
	if s < -1 {
		s = -1
	}
	v := int16(s * 32767)
	buf[i*2] = byte(v)
	buf[i*2+1] = byte(v >> 8)
}

func renderMOD(path string, sampleRate int) ([]byte, error) {
	const (
		channels     = 2
		sampleFormat = sampling.Format16BitLESigned
	)

	var features []feature.Feature
	features = append(features, feature.UseNativeSampleFormat(true))
	features = append(features, feature.IgnoreUnknownEffect{Enabled: true})
	features = append(features, feature.SongLoop{Count: 0})

	songData, songFormat, err := format.Load(path, features)
	if err != nil {
		return nil, err
	}

	var userSettings settings.UserSettings
	if err := songFormat.ConvertFeaturesToSettings(&userSettings, features); err != nil {
		return nil, err
	}

	player, err := machine.NewMachine(songData, userSettings)
	if err != nil {
		return nil, err
	}

	m := mixing.Mixer{Channels: channels}
	var pcm bytes.Buffer

	out := sampler.NewSampler(sampleRate, channels, 1.0, func(premix *output.PremixData) {
		data := m.Flatten(premix.SamplesLen, premix.Data, premix.MixerVolume, sampleFormat)
		_, _ = pcm.Write(data)
	})
	if out == nil {
		return nil, errors.New("could not create sampler")
	}

	for {
		if err := player.Advance(); err != nil {
			if errors.Is(err, song.ErrStopSong) {
				break
			}
			return nil, err
		}
		if err := player.Render(out); err != nil {
			if errors.Is(err, song.ErrStopSong) {
				break
			}
			return nil, err
		}
	}

	return pcm.Bytes(), nil
}
