package main

import (
	"math"
	"math/rand"
)

func initStars(r *rand.Rand, count int) []Star {
	stars := make([]Star, 0, count)
	for i := 0; i < count; i++ {
		size := 1.0 + r.Float64()*2.0
		speed := 0.4 + r.Float64()*1.8
		if size > 2.2 {
			speed *= 1.4
		}
		stars = append(stars, Star{
			X:     r.Float64() * screenWidth,
			Y:     r.Float64() * screenHeight,
			Speed: speed,
			Size:  size,
			Halo:  r.Float64() < 0.2,
		})
	}
	return stars
}

func updateStars(stars []Star) {
	for i := range stars {
		stars[i].Y += stars[i].Speed
		if stars[i].Y > screenHeight+4 {
			stars[i].Y = -4
			stars[i].X = rand.Float64() * screenWidth
		}
	}
}

func formationPos(side, row, col int, t float64) (float64, float64) {
	// Side: 0 = left -> right, 1 = right -> left
	offsetY := float64(row) * 48
	offsetX := float64(col) * 48 - 72

	maxOffsetX := float64((4-1)*48 - 72)
	minOffsetX := -72.0
	startX := -float64(c64SpriteWidth) - 20 - maxOffsetX
	endX := float64(screenWidth) + float64(c64SpriteWidth) + 20 - minOffsetX
	if side == 1 {
		startX, endX = endX, startX
	}
	baseY := 80 + offsetY
	p := easeInOutQuad(t)
	x := startX + (endX-startX)*p
	amp := formationWaveAmplitude
	phase := float64(row)*formationWavePhaseStep
	y := baseY + amp*math.Sin(formationWaveCycles*2*math.Pi*t+phase)
	// Column offset for spacing
	x += offsetX
	return x, y
}

func easeInOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return 1 - math.Pow(-2*t+2, 2)/2
}

func easeOutQuad(t float64) float64 {
	return 1 - (1-t)*(1-t)
}
