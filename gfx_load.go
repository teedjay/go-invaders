package main

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type AmigaFont struct {
	Glyphs map[rune]*ebiten.Image
	W      int
	H      int
}

func loadInvaderSprites() ([][]*ebiten.Image, error) {
	sheet, _, err := ebitenutil.NewImageFromFile("assets/enemies.png")
	if err != nil {
		return nil, err
	}
	frames := make([][]*ebiten.Image, invaderRows)
	for row := 0; row < invaderRows; row++ {
		frames[row] = make([]*ebiten.Image, spriteFrameCount)
		for frame := 0; frame < spriteFrameCount; frame++ {
			x0 := frame * int(invaderWidth)
			y0 := row * int(invaderHeight)
			sub := sheet.SubImage(imageRect(x0, y0, int(invaderWidth), int(invaderHeight))).(*ebiten.Image)
			frames[row][frame] = sub
		}
	}
	return frames, nil
}

func loadUFOSprites() ([]*ebiten.Image, error) {
	sheet, _, err := ebitenutil.NewImageFromFile("assets/ufo.png")
	if err != nil {
		return nil, err
	}
	frames := make([]*ebiten.Image, ufoFrameCount)
	for frame := 0; frame < ufoFrameCount; frame++ {
		x0 := frame * ufoWidth
		y0 := 0
		sub := sheet.SubImage(imageRect(x0, y0, ufoWidth, ufoHeight)).(*ebiten.Image)
		frames[frame] = sub
	}
	return frames, nil
}

func loadC64Sprites() ([][]*ebiten.Image, error) {
	sheet, _, err := ebitenutil.NewImageFromFile("assets/c64_enemies.png")
	if err != nil {
		return nil, err
	}
	frames := make([][]*ebiten.Image, c64SpriteRows)
	for row := 0; row < c64SpriteRows; row++ {
		frames[row] = make([]*ebiten.Image, c64FrameCount)
		for frame := 0; frame < c64FrameCount; frame++ {
			x0 := frame * c64SpriteWidth
			y0 := row * c64SpriteHeight
			sub := sheet.SubImage(imageRect(x0, y0, c64SpriteWidth, c64SpriteHeight)).(*ebiten.Image)
			frames[row][frame] = sub
		}
	}
	return frames, nil
}

func loadPlayerSprites() (PlayerSpriteSet, error) {
	m, err := loadImage("assets/player/player_b_m.png")
	if err != nil {
		return PlayerSpriteSet{}, err
	}
	l1, err := loadImage("assets/player/player_b_l1.png")
	if err != nil {
		return PlayerSpriteSet{}, err
	}
	l2, err := loadImage("assets/player/player_b_l2.png")
	if err != nil {
		return PlayerSpriteSet{}, err
	}
	r1, err := loadImage("assets/player/player_b_r1.png")
	if err != nil {
		return PlayerSpriteSet{}, err
	}
	r2, err := loadImage("assets/player/player_b_r2.png")
	if err != nil {
		return PlayerSpriteSet{}, err
	}
	return PlayerSpriteSet{M: m, L1: l1, L2: l2, R1: r1, R2: r2}, nil
}

func loadPlayerSuperSprites() (PlayerSpriteSet, error) {
	m, err := loadImage("assets/player/player_r_m.png")
	if err != nil {
		return PlayerSpriteSet{}, err
	}
	l1, err := loadImage("assets/player/player_r_l1.png")
	if err != nil {
		return PlayerSpriteSet{}, err
	}
	l2, err := loadImage("assets/player/player_r_l2.png")
	if err != nil {
		return PlayerSpriteSet{}, err
	}
	r1, err := loadImage("assets/player/player_r_r1.png")
	if err != nil {
		return PlayerSpriteSet{}, err
	}
	r2, err := loadImage("assets/player/player_r_r2.png")
	if err != nil {
		return PlayerSpriteSet{}, err
	}
	return PlayerSpriteSet{M: m, L1: l1, L2: l2, R1: r1, R2: r2}, nil
}

func loadImage(path string) (*ebiten.Image, error) {
	img, _, err := ebitenutil.NewImageFromFile(path)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func loadAmigaFont(path string) (*AmigaFont, error) {
	img, _, err := ebitenutil.NewImageFromFile(path)
	if err != nil {
		return nil, err
	}
	cols := 9
	rows := 4
	cellW := img.Bounds().Dx() / cols
	cellH := img.Bounds().Dy() / rows

	grid := []string{
		"ABCDEFGHI",
		"JKLMNOPQR",
		"STUVWXYZ1",
		"234567890",
	}

	glyphs := make(map[rune]*ebiten.Image)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			ch := rune(grid[r][c])
			x0 := c * cellW
			y0 := r * cellH
			sub := img.SubImage(imageRect(x0, y0, cellW, cellH)).(*ebiten.Image)
			glyphs[ch] = sub
		}
	}

	return &AmigaFont{Glyphs: glyphs, W: cellW, H: cellH}, nil
}

func imageRect(x, y, w, h int) (r image.Rectangle) {
	return image.Rect(x, y, x+w, y+h)
}
