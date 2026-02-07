package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 800
	screenHeight = 600

	playerWidth  = 48
	playerHeight = 16
	playerSpeed  = 4.0
	playerCooldownFrames = 12

	bulletWidth  = 4
	bulletHeight = 10
	bulletSpeed  = -7.0

	invaderCols = 10
	invaderRows = 4
	invaderWidth  = 32
	invaderHeight = 24
	invaderGapX   = 12
	invaderGapY   = 10
	invaderStartX = 80
	invaderStartY = 60
	invaderStepDown = 12
	invaderSpeed = 1.0

	spriteFrameCount = 4
	spriteFrameDelay = 20

	particleCount = 28
	particleLifeMax = 30
	smokeLifeMax = 24
	smokeSpawnPerFrame = 2
	shockwaveLifeMax = 18
	shockwaveSegments = 36
)

type Player struct {
	X, Y  float64
	W, H  float64
	Speed float64
	Cooldown int
}

type Invader struct {
	X, Y float64
	W, H float64
	Alive bool
	Type int
}

type Bullet struct {
	X, Y float64
	W, H float64
	Vy float64
	FromPlayer bool
	Active bool
}

type Particle struct {
	X, Y float64
	Vx, Vy float64
	Life int
	MaxLife int
	Color color.RGBA
	Size float64
	Gravity float64
}

type Shockwave struct {
	X, Y float64
	Radius float64
	Life int
	MaxLife int
	Speed float64
	Thickness float64
	Color color.RGBA
}

type Game struct {
	Player   Player
	Invaders []Invader
	Bullet   Bullet
	Score    int

	InvaderDir float64
	GameOver bool
	Win bool

	InvaderSprites [][]*ebiten.Image
	Frame int
	FrameTick int

	Particles []Particle
	Shockwaves []Shockwave
	Rand *rand.Rand
}

func NewGame() (*Game, error) {
	g := &Game{}
	g.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	g.Player = Player{
		X: (screenWidth - playerWidth) / 2,
		Y: screenHeight - 60,
		W: playerWidth,
		H: playerHeight,
		Speed: playerSpeed,
		Cooldown: 0,
	}
	g.Bullet = Bullet{
		W: bulletWidth,
		H: bulletHeight,
		Vy: bulletSpeed,
		FromPlayer: true,
		Active: false,
	}
	sprites, err := loadInvaderSprites()
	if err != nil {
		return nil, err
	}
	g.InvaderSprites = sprites

	g.Invaders = make([]Invader, 0, invaderCols*invaderRows)
	for row := 0; row < invaderRows; row++ {
		for col := 0; col < invaderCols; col++ {
			x := float64(invaderStartX + col*(invaderWidth+invaderGapX))
			y := float64(invaderStartY + row*(invaderHeight+invaderGapY))
			g.Invaders = append(g.Invaders, Invader{
				X: x,
				Y: y,
				W: invaderWidth,
				H: invaderHeight,
				Alive: true,
				Type: row,
			})
		}
	}
	g.InvaderDir = 1
	return g, nil
}

func (g *Game) Update() error {
	if g.GameOver || g.Win {
		return nil
	}

	// Movement
	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.Player.X -= g.Player.Speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.Player.X += g.Player.Speed
	}
	if g.Player.X < 0 {
		g.Player.X = 0
	}
	if g.Player.X+g.Player.W > screenWidth {
		g.Player.X = screenWidth - g.Player.W
	}

	// Fire
	if g.Player.Cooldown > 0 {
		g.Player.Cooldown--
	}
	if ebiten.IsKeyPressed(ebiten.KeySpace) && g.Player.Cooldown == 0 && !g.Bullet.Active {
		g.Bullet.Active = true
		g.Bullet.X = g.Player.X + g.Player.W/2 - g.Bullet.W/2
		g.Bullet.Y = g.Player.Y - g.Bullet.H
		g.Player.Cooldown = playerCooldownFrames
	}

	// Bullet update
	if g.Bullet.Active {
		g.Bullet.Y += g.Bullet.Vy
		if g.Bullet.Y+g.Bullet.H < 0 {
			g.Bullet.Active = false
		}
		// Smoke trail
		for i := 0; i < smokeSpawnPerFrame; i++ {
			sx := g.Bullet.X + g.Bullet.W/2 + (g.Rand.Float64()-0.5)*4
			sy := g.Bullet.Y + g.Bullet.H + g.Rand.Float64()*2
			g.spawnSmoke(sx, sy)
		}
	}

	// Invader movement
	minX, maxX := g.invaderBounds()
	if maxX+invaderSpeed > screenWidth || minX-invaderSpeed < 0 {
		g.InvaderDir *= -1
		for i := range g.Invaders {
			if g.Invaders[i].Alive {
				g.Invaders[i].Y += invaderStepDown
			}
		}
	}
	for i := range g.Invaders {
		if g.Invaders[i].Alive {
			g.Invaders[i].X += g.InvaderDir * invaderSpeed
		}
	}

	// Collisions: bullet vs invaders
	if g.Bullet.Active {
		for i := range g.Invaders {
			inv := &g.Invaders[i]
			if !inv.Alive {
				continue
			}
			if rectsOverlap(g.Bullet.X, g.Bullet.Y, g.Bullet.W, g.Bullet.H, inv.X, inv.Y, inv.W, inv.H) {
				inv.Alive = false
				g.Bullet.Active = false
				g.spawnExplosion(inv.X+inv.W/2, inv.Y+inv.H/2)
				g.Score += 10
				break
			}
		}
	}

	// Win/Lose check
	aliveCount := 0
	for i := range g.Invaders {
		if g.Invaders[i].Alive {
			aliveCount++
			if g.Invaders[i].Y+g.Invaders[i].H >= g.Player.Y {
				g.GameOver = true
			}
		}
	}
	if aliveCount == 0 {
		g.Win = true
	}

	// Particles update
	if len(g.Particles) > 0 {
		next := g.Particles[:0]
		for i := range g.Particles {
			p := g.Particles[i]
			p.X += p.Vx
			p.Y += p.Vy
			p.Vy += p.Gravity
			p.Life--
			if p.Life > 0 {
				next = append(next, p)
			}
		}
		g.Particles = next
	}

	// Shockwaves update
	if len(g.Shockwaves) > 0 {
		next := g.Shockwaves[:0]
		for i := range g.Shockwaves {
			s := g.Shockwaves[i]
			s.Radius += s.Speed
			s.Life--
			if s.Life > 0 {
				next = append(next, s)
			}
		}
		g.Shockwaves = next
	}

	// Animation tick
	g.FrameTick++
	if g.FrameTick >= spriteFrameDelay {
		g.FrameTick = 0
		g.Frame = (g.Frame + 1) % spriteFrameCount
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 12, G: 16, B: 20, A: 255})

	// Player
	ebitenutil.DrawRect(screen, g.Player.X, g.Player.Y, g.Player.W, g.Player.H, color.RGBA{R: 40, G: 200, B: 120, A: 255})

	// Invaders
	for i := range g.Invaders {
		if !g.Invaders[i].Alive {
			continue
		}
		inv := &g.Invaders[i]
		if inv.Type >= 0 && inv.Type < len(g.InvaderSprites) && g.Frame < len(g.InvaderSprites[inv.Type]) {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(inv.X, inv.Y)
			screen.DrawImage(g.InvaderSprites[inv.Type][g.Frame], op)
		} else {
			ebitenutil.DrawRect(screen, inv.X, inv.Y, inv.W, inv.H, color.RGBA{R: 200, G: 70, B: 60, A: 255})
		}
	}

	// Bullet
	if g.Bullet.Active {
		// Glow layers
		ebitenutil.DrawRect(screen, g.Bullet.X-3, g.Bullet.Y-4, g.Bullet.W+6, g.Bullet.H+8, color.RGBA{R: 120, G: 200, B: 255, A: 60})
		ebitenutil.DrawRect(screen, g.Bullet.X-1, g.Bullet.Y-2, g.Bullet.W+2, g.Bullet.H+4, color.RGBA{R: 180, G: 230, B: 255, A: 120})
		ebitenutil.DrawRect(screen, g.Bullet.X, g.Bullet.Y, g.Bullet.W, g.Bullet.H, color.RGBA{R: 240, G: 240, B: 255, A: 255})
	}

	// Shockwaves
	for i := range g.Shockwaves {
		s := &g.Shockwaves[i]
		alpha := uint8(200 * s.Life / s.MaxLife)
		c := color.RGBA{R: s.Color.R, G: s.Color.G, B: s.Color.B, A: alpha}
		for seg := 0; seg < shockwaveSegments; seg++ {
			ang := (float64(seg) / float64(shockwaveSegments)) * 2 * math.Pi
			r := s.Radius + math.Sin(float64(seg))*s.Thickness
			x := s.X + math.Cos(ang)*r
			y := s.Y + math.Sin(ang)*r
			ebitenutil.DrawRect(screen, x, y, 2, 2, c)
		}
	}

	// Particles
	for i := range g.Particles {
		p := &g.Particles[i]
		alpha := uint8(255 * p.Life / p.MaxLife)
		c := color.RGBA{R: p.Color.R, G: p.Color.G, B: p.Color.B, A: alpha}
		ebitenutil.DrawRect(screen, p.X, p.Y, p.Size, p.Size, c)
	}

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d", g.Score), 10, 10)
	if g.Win {
		ebitenutil.DebugPrintAt(screen, "WIN", screenWidth/2-20, screenHeight/2-10)
	}
	if g.GameOver {
		ebitenutil.DebugPrintAt(screen, "GAME OVER", screenWidth/2-40, screenHeight/2-10)
	}
}

func (g *Game) Layout(_, _ int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) invaderBounds() (minX float64, maxX float64) {
	minX = screenWidth
	maxX = 0
	for i := range g.Invaders {
		if !g.Invaders[i].Alive {
			continue
		}
		if g.Invaders[i].X < minX {
			minX = g.Invaders[i].X
		}
		right := g.Invaders[i].X + g.Invaders[i].W
		if right > maxX {
			maxX = right
		}
	}
	if maxX == 0 && minX == screenWidth {
		minX = 0
		maxX = 0
	}
	return minX, maxX
}

func rectsOverlap(ax, ay, aw, ah, bx, by, bw, bh float64) bool {
	return ax < bx+bw && ax+aw > bx && ay < by+bh && ay+ah > by
}

func (g *Game) spawnExplosion(x, y float64) {
	colors := []color.RGBA{
		{R: 255, G: 220, B: 100, A: 255},
		{R: 255, G: 140, B: 60, A: 255},
		{R: 255, G: 80, B: 120, A: 255},
		{R: 120, G: 200, B: 255, A: 255},
	}
	for i := 0; i < particleCount; i++ {
		angle := g.Rand.Float64() * 2 * 3.1415926
		speed := 1.2 + g.Rand.Float64()*3.2
		vx := math.Cos(angle) * speed
		vy := math.Sin(angle) * speed
		c := colors[g.Rand.Intn(len(colors))]
		g.Particles = append(g.Particles, Particle{
			X: x,
			Y: y,
			Vx: vx,
			Vy: vy,
			Life: particleLifeMax,
			MaxLife: particleLifeMax,
			Color: c,
			Size: 2 + g.Rand.Float64()*2,
			Gravity: 0.05,
		})
	}

	g.Shockwaves = append(g.Shockwaves, Shockwave{
		X: x,
		Y: y,
		Radius: 4,
		Life: shockwaveLifeMax,
		MaxLife: shockwaveLifeMax,
		Speed: 3.2,
		Thickness: 1.5,
		Color: color.RGBA{R: 180, G: 220, B: 255, A: 255},
	})
}

func (g *Game) spawnSmoke(x, y float64) {
	c := []color.RGBA{
		{R: 120, G: 140, B: 160, A: 255},
		{R: 90, G: 110, B: 130, A: 255},
		{R: 160, G: 180, B: 200, A: 255},
	}[g.Rand.Intn(3)]
	vx := (g.Rand.Float64()-0.5) * 0.6
	vy := 0.4 + g.Rand.Float64()*0.6
	g.Particles = append(g.Particles, Particle{
		X: x,
		Y: y,
		Vx: vx,
		Vy: vy,
		Life: smokeLifeMax,
		MaxLife: smokeLifeMax,
		Color: c,
		Size: 2 + g.Rand.Float64()*2,
		Gravity: 0.02,
	})
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Go Invaders")
	game, err := NewGame()
	if err != nil {
		log.Fatal(err)
	}
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func loadInvaderSprites() ([][]*ebiten.Image, error) {
	sheet, _, err := ebitenutil.NewImageFromFile("/Users/thorej/opt/codex/go-invaders/assets/enemies.png")
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

func imageRect(x, y, w, h int) (r image.Rectangle) {
	return image.Rect(x, y, x+w, y+h)
}
