package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

func (g *Game) updateStartScreen() bool {
	if g.StartState == startStateDone {
		return false
	}
	g.StartTick++
	if g.StartState == startStateScreen && g.StartTick >= 2520 {
		g.StartTick = 0
	}
	if g.StartState == startStateScreen {
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.StartState = startStateFading
			g.StartFadeTick = 0
		}
	} else if g.StartState == startStateFading {
		g.StartFadeTick++
		fadeT := float64(g.StartFadeTick) / float64(startFadeFrames)
		if fadeT > 1 {
			fadeT = 1
		}
		g.setMusicVolume(0.3 * (1.0 - fadeT))
		if g.StartFadeTick >= startFadeFrames {
			g.stopMusic()
			g.StartState = startStateDone
			g.resetLevelState()
		}
	}
	return true
}

func (g *Game) drawStartScreen(screen *ebiten.Image) bool {
	if g.StartState == startStateDone {
		return false
	}
	drawStartImage(screen, g.StartImage)

	alpha := 0.0
	if g.StartTick >= 120 && g.StartTick < 180 {
		alpha = float64(g.StartTick-120) / 60.0
	} else if g.StartTick >= 180 && g.StartTick < 1800 {
		alpha = 1.0
	} else if g.StartTick >= 1800 && g.StartTick < 1920 {
		alpha = 1.0 - float64(g.StartTick-1800)/120.0
	}

	var waveAmp float64
	if g.StartTick >= 300 && g.StartTick < 600 {
		t := float64(g.StartTick-300) / 300.0
		waveAmp = 50 * t
	} else if g.StartTick >= 600 && g.StartTick < 1200 {
		waveAmp = 50
	} else if g.StartTick >= 1200 && g.StartTick < 1500 {
		t := float64(g.StartTick-1200) / 300.0
		waveAmp = 50 * (1.0 - t)
	}
	drawAmigaText(screen, g.StartFont, "GO INVADERS", screenWidth/2, int(float64(screenHeight)*0.2), alpha, waveAmp, g.StartTick)

	if g.StartTick >= startPromptDelayFrames {
		glow := 0.5 + 0.5*math.Sin(float64(g.StartTick)*0.08)
		drawGlowText(screen, "PRESS FIRE TO START", screenWidth/2, int(float64(screenHeight)*0.65), 3.0, glow)
	}

	if g.StartState == startStateFading {
		fadeAlpha := easeInOutQuad(float64(g.StartFadeTick) / float64(startFadeFrames))
		if fadeAlpha > 1 {
			fadeAlpha = 1
		}
		drawFadeOverlay(screen, 0, 0, 0, fadeAlpha)
	}
	return true
}
