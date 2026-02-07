package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var fadePixel *ebiten.Image

func drawHugeText(screen *ebiten.Image, text string, centerX, centerY int, scaleFactor float64) {
	scale := 4
	baseW := len(text) * 7
	baseH := 16
	img := ebiten.NewImage(baseW, baseH)
	img.Fill(color.RGBA{0, 0, 0, 0})
	ebitenutil.DebugPrintAt(img, text, 0, 4)

	scaleF := float64(scale) * scaleFactor
	w := float64(baseW) * scaleF
	h := float64(baseH) * scaleF
	x := float64(centerX) - w/2
	y := float64(centerY) - h/2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scaleF, scaleF)
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
}

func drawGlowText(screen *ebiten.Image, text string, centerX, centerY int, scaleFactor float64, glow float64) {
	baseW := len(text) * 7
	baseH := 16
	img := ebiten.NewImage(baseW, baseH)
	img.Fill(color.RGBA{0, 0, 0, 0})
	ebitenutil.DebugPrintAt(img, text, 0, 4)

	scaleF := scaleFactor
	w := float64(baseW) * scaleF
	h := float64(baseH) * scaleF
	x := float64(centerX) - w/2
	y := float64(centerY) - h/2

	// Main text color fades white -> yellow
	mainOp := &ebiten.DrawImageOptions{}
	mainOp.GeoM.Scale(scaleF, scaleF)
	mainOp.GeoM.Translate(x, y)
	r := 1.0
	gc := 1.0
	b := 1.0 - 0.7*glow
	mainOp.ColorM.Scale(r, gc, b, 0.95)
	screen.DrawImage(img, mainOp)
}

func drawFadeOverlay(screen *ebiten.Image, r, g, b, alpha float64) {
	if fadePixel == nil {
		fadePixel = ebiten.NewImage(1, 1)
		fadePixel.Fill(color.RGBA{R: 255, G: 255, B: 255, A: 255})
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(screenWidth), float64(screenHeight))
	op.ColorM.Scale(r, g, b, alpha)
	screen.DrawImage(fadePixel, op)
}

func drawStartImage(screen *ebiten.Image, img *ebiten.Image) {
	if img == nil {
		return
	}
	w := float64(img.Bounds().Dx())
	h := float64(img.Bounds().Dy())
	sw := float64(screenWidth)
	sh := float64(screenHeight)
	if w == 0 || h == 0 {
		return
	}
	scale := sw / w
	if sh/h > scale {
		scale = sh / h
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	x := (sw - w*scale) / 2
	y := (sh - h*scale) / 2
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
}

func drawStarfield(screen *ebiten.Image, stars []Star) {
	for _, s := range stars {
		if s.Halo {
			ebitenutil.DrawRect(screen, s.X-1, s.Y-1, s.Size+2, s.Size+2, color.RGBA{R: 120, G: 140, B: 180, A: 80})
		}
		ebitenutil.DrawRect(screen, s.X, s.Y, s.Size, s.Size, color.RGBA{R: 220, G: 230, B: 255, A: 200})
	}
}

func drawHUD(screen *ebiten.Image, g *Game) {
	scale := 3.0
	left := fmt.Sprintf("SCORE: %d", g.Score)
	center := fmt.Sprintf("LEVEL %d", g.Level)
	right := fmt.Sprintf("LIVES: %d", g.Lives)

	drawHUDText(screen, left, 10, 8, scale)

	centerWidth := hudTextWidth(center, scale)
	drawHUDText(screen, center, screenWidth/2-centerWidth/2, 8, scale)

	rightWidth := hudTextWidth(right, scale)
	drawHUDText(screen, right, screenWidth-10-rightWidth, 8, scale)
}

func hudTextWidth(text string, scale float64) int {
	baseW := len(text) * 7
	return int(float64(baseW) * scale)
}

func drawHUDText(screen *ebiten.Image, text string, x, y int, scale float64) {
	baseW := len(text) * 7
	baseH := 16
	img := ebiten.NewImage(baseW, baseH)
	img.Fill(color.RGBA{0, 0, 0, 0})
	ebitenutil.DebugPrintAt(img, text, 0, 4)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, op)
}

func drawAmigaText(screen *ebiten.Image, font *AmigaFont, text string, centerX, y int, alpha float64, waveAmp float64, tick int) {
	if font == nil || font.W == 0 {
		return
	}
	scale := 2.0
	spacing := 1
	totalW := 0
	for _, r := range text {
		if r == ' ' || font.Glyphs[r] == nil {
			totalW += int(float64(font.W+spacing) * scale)
			continue
		}
		totalW += int(float64(font.W+spacing) * scale)
	}
	if totalW > 0 {
		totalW -= int(float64(spacing) * scale)
	}
	x := centerX - totalW/2
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}
	for _, r := range text {
		if r == ' ' || font.Glyphs[r] == nil {
			x += int(float64(font.W+spacing) * scale)
			continue
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		offsetY := 0.0
		if waveAmp > 0 {
			offsetY = math.Sin(float64(tick)*0.12+float64(x)*0.03) * waveAmp
		}
		op.GeoM.Translate(float64(x), float64(y)+offsetY)
		op.ColorM.Scale(1, 1, 1, alpha)
		screen.DrawImage(font.Glyphs[r], op)
		x += int(float64(font.W+spacing) * scale)
	}
}

func (g *Game) playerSpriteImage() *ebiten.Image {
	sets := g.PlayerSprites
	if g.Player.SuperFrames > 0 {
		blinkWindow := g.Player.SuperFrames > playerSuperDurationFrames-playerSuperBlinkFirstFrames ||
			g.Player.SuperFrames <= playerSuperBlinkLastFrames
		if blinkWindow && (g.Player.SuperFrames/playerSuperBlinkPeriod)%2 == 1 {
			sets = g.PlayerSprites
		} else {
			sets = g.PlayerSuperSprites
		}
	}
	if sets.M == nil {
		return nil
	}
	if g.Player.AnimStep == 0 {
		return sets.M
	}
	if g.Player.AnimDir < 0 {
		if g.Player.AnimStep == 1 {
			return sets.L1
		}
		return sets.L2
	}
	if g.Player.AnimDir > 0 {
		if g.Player.AnimStep == 1 {
			return sets.R1
		}
		return sets.R2
	}
	return sets.M
}
