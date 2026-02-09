package main

import (
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 800
	screenHeight = 600

	playerWidth          = 48
	playerHeight         = 16
	playerSpeed          = 4.0
	playerCooldownFrames = 12

	bulletWidth  = 4
	bulletHeight = 10
	bulletSpeed  = -7.0

	invaderCols                = 10
	invaderRows                = 4
	invaderWidth               = 32
	invaderHeight              = 24
	invaderGapX                = 12
	invaderGapY                = 10
	invaderStartX              = 80
	invaderStartY              = 60
	invaderStepDown            = 12
	invaderSpeed               = 1.0
	invaderSpeedIncreaseFactor = 0.2

	spriteFrameCount = 4
	spriteFrameDelay = 20
	deathAnimFrames  = 18

	particleCount      = 28
	particleLifeMax    = 30
	smokeLifeMax       = 24
	smokeSpawnPerFrame = 2
	shockwaveLifeMax   = 18
	shockwaveSegments  = 36

	ufoWidth       = 64
	ufoHeight      = 32
	ufoFrameCount  = 8
	ufoFrameDelay  = 8
	ufoSpawnFrames = 600
	ufoSpeed       = 5.5

	c64SpriteWidth              = 32
	c64SpriteHeight             = 24
	c64SpriteRows               = 8
	c64FrameCount               = 4
	c64FrameDelay               = 8
	formationDurationFrames     = 300
	formationSpawnFrames        = 360
	formationInitialDelayFrames = 300
	formationWaveAmplitude      = 60.0
	formationWaveCycles         = 2.0
	formationWavePhaseStep      = 0.6

	formationGravity                = 0.35
	formationShotSpeedY             = -6.0
	formationShotSpeedXMin          = -2.5
	formationShotSpeedXMax          = 2.5
	hugeExplosionParticleMultiplier = 4
	hugeExplosionShockwaveScale     = 1.8

	completeTextDurationFrames = 90
	completeTextStartY         = -80.0
	killAllIntervalFrames      = 3
	fadeWhiteFrames            = 30
	fadeBlackFrames            = 30
	startPromptDelayFrames     = 600
	startFadeFrames            = 120
	gameOverZoomFrames         = 120
	gameOverShipFadeFrames     = 120
	gameOverWaitFrames         = 900
	gameOverFadeFrames         = 120

	playerAnimDelay             = 6
	playerSuperDurationFrames   = 480
	playerSuperBlinkFirstFrames = 60
	playerSuperBlinkLastFrames  = 120
	playerSuperBlinkPeriod      = 6
)

type Player struct {
	X, Y        float64
	W, H        float64
	Speed       float64
	Cooldown    int
	SuperFrames int
	AnimDir     int
	AnimStep    int
	AnimTick    int
}

type Star struct {
	X, Y  float64
	Speed float64
	Size  float64
	Halo  bool
}

type Invader struct {
	X, Y      float64
	W, H      float64
	Alive     bool
	Type      int
	Dying     bool
	DeathTick int
}

type Bullet struct {
	X, Y       float64
	BaseX      float64
	W, H       float64
	Vy         float64
	Age        int
	Phase      float64
	FromPlayer bool
	Active     bool
	ID         int
	Wheee      *audio.Player
	StartY     float64
}

type Particle struct {
	X, Y    float64
	Vx, Vy  float64
	Life    int
	MaxLife int
	Color   color.RGBA
	Size    float64
	Gravity float64
}

type Shockwave struct {
	X, Y      float64
	Radius    float64
	Life      int
	MaxLife   int
	Speed     float64
	Thickness float64
	Color     color.RGBA
}

type UFO struct {
	X, Y        float64
	Vx          float64
	Active      bool
	Frame       int
	FrameTick   int
	Crashing    bool
	CrashAngle  float64
	CrashRadius float64
	CrashVy     float64
}

type FlyEnemy struct {
	X, Y         float64
	PrevX, PrevY float64
	Angle        float64
	Frame        int
	FrameTick    int
	Index        int
	Side         int
	SpriteIndex  int
	Shot         bool
	Vx           float64
	Vy           float64
}

type Formation struct {
	Active   bool
	Tick     int
	Duration int
	Enemies  [16]FlyEnemy
	Side     int
}

type PlayerSpriteSet struct {
	M  *ebiten.Image
	L1 *ebiten.Image
	L2 *ebiten.Image
	R1 *ebiten.Image
	R2 *ebiten.Image
}

type Game struct {
	Player   Player
	Invaders []Invader
	Bullets  []Bullet
	Score    int

	InvaderDir            float64
	GameOver              bool
	Win                   bool
	InvaderSpeed          float64
	InvaderSpeedThreshold int
	InvaderSpeedIncrease  float64

	InvaderSprites [][]*ebiten.Image
	Frame          int
	FrameTick      int

	Particles  []Particle
	Shockwaves []Shockwave
	Rand       *rand.Rand

	UFOSprites   []*ebiten.Image
	UFO          UFO
	UFOSpawnTick int

	C64Sprites         [][]*ebiten.Image
	Formation          Formation
	FormationSpawnTick int
	C64Frame           int
	C64FrameTick       int

	PlayerSprites      PlayerSpriteSet
	PlayerSuperSprites PlayerSpriteSet

	LevelComplete     bool
	CompleteTick      int
	Level             int
	LevelState        int
	FadeTick          int
	Lives             int
	GameOverTick      int
	GameOverTriggered bool
	GameOverWaitTick  int
	GameOverFadeTick  int
	GameOverFading    bool

	KillAllActive bool
	KillAllTick   int
	KillAllOrder  []int

	AudioCtx     *audio.Context
	Sfx          SFX
	NextBulletID int
	UFOLoop      *audio.Player
	SuperLoop    *audio.Player
	MusicPlayer  *audio.Player

	StartImage    *ebiten.Image
	StartState    int
	StartTick     int
	StartFadeTick int
	StartFont     *AmigaFont

	Stars []Star
}

const (
	levelStatePlaying = iota
	levelStateCompleteWait
	levelStateFadeWhite
	levelStateFadeBlack
)

const (
	startStateScreen = iota
	startStateFading
	startStateDone
)

func NewGame() (*Game, error) {
	g := &Game{}
	g.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	g.Level = 1
	g.LevelState = levelStatePlaying
	g.Lives = 3
	g.AudioCtx = audio.NewContext(audioSampleRate)
	g.Sfx = SFX{
		Swoosh:    genSwoosh(audioSampleRate, 0.14),
		Wheee:     genWheee(audioSampleRate, 1.2),
		Explosion: genExplosion(audioSampleRate, 0.45),
		UFOWooo:   genUFOWooo(audioSampleRate, 1.4),
		SuperDrum: genSuperDrum(audioSampleRate, 0.5),
	}
	g.StartState = startStateScreen
	g.StartTick = 0
	g.StartFadeTick = 0
	g.StartImage, _ = loadImage("assets/image.png")
	g.StartFont, _ = loadAmigaFont("assets/amiga_font.png")
	g.startMusic("music/codex_amiga_mstb.mod")
	g.Stars = initStars(g.Rand, 160)
	g.Player = Player{
		X:        (screenWidth - playerWidth) / 2,
		Y:        screenHeight - 60,
		W:        playerWidth,
		H:        playerHeight,
		Speed:    playerSpeed,
		Cooldown: 0,
	}
	g.Bullets = make([]Bullet, 0, 5)
	sprites, err := loadInvaderSprites()
	if err != nil {
		return nil, err
	}
	g.InvaderSprites = sprites
	ufoSprites, err := loadUFOSprites()
	if err != nil {
		return nil, err
	}
	g.UFOSprites = ufoSprites
	g.UFOSpawnTick = ufoSpawnFrames

	c64Sprites, err := loadC64Sprites()
	if err != nil {
		return nil, err
	}
	g.C64Sprites = c64Sprites
	g.FormationSpawnTick = formationInitialDelayFrames
	g.Formation = Formation{Active: false, Tick: 0, Duration: formationDurationFrames}

	playerSprites, err := loadPlayerSprites()
	if err != nil {
		return nil, err
	}
	playerSuperSprites, err := loadPlayerSuperSprites()
	if err != nil {
		return nil, err
	}
	g.PlayerSprites = playerSprites
	g.PlayerSuperSprites = playerSuperSprites

	g.resetLevelState()
	return g, nil
}

func (g *Game) Update() error {
	if g.GameOver {
		if !g.GameOverTriggered {
			g.GameOverTriggered = true
			g.GameOverTick = 0
			g.GameOverWaitTick = 0
			g.GameOverFadeTick = 0
			g.GameOverFading = false
			g.stopUFOLoop()
			g.spawnGameOverExplosion(g.Player.X+g.Player.W/2, g.Player.Y+g.Player.H/2)
		}
		g.GameOverTick++
		if !g.GameOverFading {
			if g.GameOverTick >= gameOverZoomFrames {
				g.GameOverWaitTick++
				if g.GameOverWaitTick >= gameOverWaitFrames || ebiten.IsKeyPressed(ebiten.KeySpace) {
					g.GameOverFading = true
					g.GameOverFadeTick = 0
				}
			}
		} else {
			g.GameOverFadeTick++
			if g.GameOverFadeTick >= gameOverFadeFrames {
				g.resetToStartScreen()
			}
		}
		if g.StartState == startStateDone {
			updateStars(g.Stars)
		}
		g.updateEffects()
		return nil
	}

	// Start screen flow
	if g.updateStartScreen() {
		return nil
	}

	// Manual end-game trigger
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		g.GameOver = true
	}

	// Fade transitions
	if g.LevelState == levelStateFadeWhite {
		g.FadeTick++
		if g.FadeTick >= fadeWhiteFrames {
			g.LevelState = levelStateFadeBlack
			g.FadeTick = 0
		}
		return nil
	}
	if g.LevelState == levelStateFadeBlack {
		g.FadeTick++
		if g.FadeTick >= fadeBlackFrames {
			g.Level++
			g.resetLevelState()
		}
		return nil
	}

	// Starfield update (gameplay + level complete + fades)
	if g.StartState == startStateDone {
		updateStars(g.Stars)
	}

	// Hidden: kill all regular aliens (F)
	if ebiten.IsKeyPressed(ebiten.KeyF) && !g.KillAllActive {
		g.KillAllOrder = g.KillAllOrder[:0]
		for i := range g.Invaders {
			if g.Invaders[i].Alive {
				g.KillAllOrder = append(g.KillAllOrder, i)
			}
		}
		if len(g.KillAllOrder) > 0 {
			g.Rand.Shuffle(len(g.KillAllOrder), func(i, j int) {
				g.KillAllOrder[i], g.KillAllOrder[j] = g.KillAllOrder[j], g.KillAllOrder[i]
			})
			g.KillAllActive = true
			g.KillAllTick = 0
		}
	}

	// Movement
	moveSpeed := g.Player.Speed
	if g.Player.SuperFrames > 0 {
		moveSpeed *= 2
	}
	movingLeft := ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA)
	movingRight := ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD)
	if movingLeft {
		g.Player.X -= moveSpeed
	}
	if movingRight {
		g.Player.X += moveSpeed
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
	if !g.LevelComplete && ebiten.IsKeyPressed(ebiten.KeySpace) && g.Player.Cooldown == 0 && len(g.Bullets) < cap(g.Bullets) {
		b := Bullet{
			W:          bulletWidth,
			H:          bulletHeight,
			Vy:         bulletSpeed,
			FromPlayer: true,
			Active:     true,
			Age:        0,
			Phase:      g.Rand.Float64() * 2 * math.Pi,
			ID:         g.NextBulletID,
		}
		g.NextBulletID++
		b.BaseX = g.Player.X + g.Player.W/2 - b.W/2
		b.X = b.BaseX
		b.Y = g.Player.Y + g.Player.H
		b.StartY = b.Y
		b.Wheee = g.playWheee()
		g.Bullets = append(g.Bullets, b)
		g.Player.Cooldown = playerCooldownFrames
		g.playSfx(g.Sfx.Swoosh, sfxVolume)
	}

	// Bullet update
	if len(g.Bullets) > 0 {
		nextBullets := g.Bullets[:0]
		for i := range g.Bullets {
			b := g.Bullets[i]
			if !b.Active {
				g.stopWheee(b.Wheee)
				b.Wheee = nil
				continue
			}
			b.Age++
			t := float64(b.Age) / 30.0
			if t > 1 {
				t = 1
			}
			speedMultiplier := 1.0
			if g.Player.SuperFrames > 0 {
				speedMultiplier = 2.0
			}
			vy := b.Vy * (1 + t) * speedMultiplier
			b.Y += vy
			wobble := math.Sin(b.Phase+float64(b.Age)*0.4) * 3.5
			b.X = b.BaseX + wobble
			if b.Y+b.H < 0 {
				g.stopWheee(b.Wheee)
				b.Wheee = nil
				continue
			}
			g.updateWheee(b.Wheee, b.StartY, b.Y)
			// Smoke trail
			for s := 0; s < smokeSpawnPerFrame; s++ {
				sx := b.X + b.W/2 + (g.Rand.Float64()-0.5)*4
				sy := b.Y + b.H + g.Rand.Float64()*2
				g.spawnSmoke(sx, sy)
			}
			nextBullets = append(nextBullets, b)
		}
		g.Bullets = nextBullets
	}

	// Invader movement
	minX, maxX := g.invaderBounds()
	if maxX+g.InvaderSpeed > screenWidth || minX-g.InvaderSpeed < 0 {
		g.InvaderDir *= -1
		for i := range g.Invaders {
			if g.Invaders[i].Alive {
				g.Invaders[i].Y += invaderStepDown
			}
		}
	}
	for i := range g.Invaders {
		if g.Invaders[i].Alive {
			g.Invaders[i].X += g.InvaderDir * g.InvaderSpeed
		}
	}

	// UFO spawn + movement
	if !g.UFO.Active {
		g.stopUFOLoop()
		g.UFOSpawnTick--
		if g.UFOSpawnTick <= 0 {
			fromLeft := g.Rand.Intn(2) == 0
			if fromLeft {
				g.UFO = UFO{X: -ufoWidth, Y: 24, Vx: ufoSpeed, Active: true}
			} else {
				g.UFO = UFO{X: screenWidth + ufoWidth, Y: 24, Vx: -ufoSpeed, Active: true}
			}
			g.startUFOLoop()
		}
	} else {
		if g.UFO.Crashing {
			g.stopUFOLoop()
			g.UFO.CrashAngle += 0.35
			g.UFO.CrashRadius += 0.08
			// Keep horizontal velocity while spiraling downward
			g.UFO.X += g.UFO.Vx
			g.UFO.X += math.Cos(g.UFO.CrashAngle) * g.UFO.CrashRadius
			g.UFO.Y += g.UFO.CrashVy
			g.UFO.Y += math.Sin(g.UFO.CrashAngle) * g.UFO.CrashRadius
			g.UFO.CrashVy += 0.12
			for i := 0; i < 6; i++ {
				g.spawnSmokeBlack(g.UFO.X+float64(ufoWidth)/2, g.UFO.Y+float64(ufoHeight)/2)
			}
			if g.UFO.Y > screenHeight+ufoHeight {
				g.UFO.Active = false
				g.UFOSpawnTick = ufoSpawnFrames
			}
		} else {
			g.startUFOLoop()
			g.UFO.X += g.UFO.Vx
		}
		g.UFO.FrameTick++
		if g.UFO.FrameTick >= ufoFrameDelay {
			g.UFO.FrameTick = 0
			g.UFO.Frame = (g.UFO.Frame + 1) % ufoFrameCount
		}
		if !g.UFO.Crashing && (g.UFO.X < -ufoWidth*2 || g.UFO.X > screenWidth+ufoWidth*2) {
			g.UFO.Active = false
			g.UFOSpawnTick = ufoSpawnFrames
			g.stopUFOLoop()
		}
	}

	// Player animation and super mode timer
	desiredDir := 0
	if movingLeft && !movingRight {
		desiredDir = -1
	} else if movingRight && !movingLeft {
		desiredDir = 1
	}
	g.Player.AnimTick++
	if g.Player.AnimTick >= playerAnimDelay {
		g.Player.AnimTick = 0
		if desiredDir != 0 {
			g.Player.AnimDir = desiredDir
			if g.Player.AnimStep == 0 {
				g.Player.AnimStep = 1
			} else if g.Player.AnimStep == 1 {
				g.Player.AnimStep = 2
			} else {
				g.Player.AnimStep = 1
			}
		} else if g.Player.AnimStep > 0 {
			g.Player.AnimStep--
		}
	}
	if g.Player.SuperFrames > 0 {
		g.Player.SuperFrames--
		g.startSuperLoop()
	} else {
		g.stopSuperLoop()
	}

	// C64 formation spawn + movement
	if !g.Formation.Active {
		g.FormationSpawnTick--
		if g.FormationSpawnTick <= 0 {
			g.startFormation()
		}
	} else {
		g.Formation.Tick++
		t := float64(g.Formation.Tick) / float64(g.Formation.Duration)
		if t >= 1.0 {
			g.Formation.Active = false
			g.FormationSpawnTick = formationSpawnFrames
		} else {
			for i := range g.Formation.Enemies {
				e := &g.Formation.Enemies[i]
				e.PrevX = e.X
				e.PrevY = e.Y
				if e.Shot {
					e.X += e.Vx
					e.Y += e.Vy
					e.Vy += formationGravity
					e.Angle = math.Atan2(e.Y-e.PrevY, e.X-e.PrevX)
					if e.Y > screenHeight+float64(c64SpriteHeight) {
						e.Shot = false
						e.X = -1000
						e.Y = -1000
					}
					continue
				}
				row := e.Index / 4
				col := e.Index % 4
				e.X, e.Y = formationPos(e.Side, row, col, t)
				e.Angle = math.Atan2(e.Y-e.PrevY, e.X-e.PrevX)
			}
			g.C64FrameTick++
			if g.C64FrameTick >= c64FrameDelay {
				g.C64FrameTick = 0
				g.C64Frame = (g.C64Frame + 1) % c64FrameCount
			}
		}
	}

	// Collisions: bullet vs invaders
	if len(g.Bullets) > 0 {
		nextBullets := g.Bullets[:0]
		for bi := range g.Bullets {
			b := g.Bullets[bi]
			if g.Formation.Active {
				for i := range g.Formation.Enemies {
					e := &g.Formation.Enemies[i]
					if e.Shot {
						continue
					}
					if rectsOverlap(b.X, b.Y, b.W, b.H, e.X, e.Y, c64SpriteWidth, c64SpriteHeight) {
						e.Shot = true
						e.Vy = formationShotSpeedY
						e.Vx = formationShotSpeedXMin + g.Rand.Float64()*(formationShotSpeedXMax-formationShotSpeedXMin)
						g.spawnHugeExplosion(e.X+float64(c64SpriteWidth)/2, e.Y+float64(c64SpriteHeight)/2)
						g.Score += 10
						b.Active = false
						g.stopWheee(b.Wheee)
						b.Wheee = nil
						break
					}
				}
				if !b.Active {
					continue
				}
			}
			if g.UFO.Active && !g.UFO.Crashing {
				if rectsOverlap(b.X, b.Y, b.W, b.H, g.UFO.X, g.UFO.Y, ufoWidth, ufoHeight) {
					g.spawnBigExplosion(g.UFO.X+float64(ufoWidth)/2, g.UFO.Y+float64(ufoHeight)/2)
					g.UFO.Crashing = true
					g.UFO.CrashAngle = 0
					g.UFO.CrashRadius = 0.6
					g.UFO.CrashVy = 1.2
					g.Player.SuperFrames = playerSuperDurationFrames
					continue
				}
			}
			hit := false
			for i := range g.Invaders {
				inv := &g.Invaders[i]
				if !inv.Alive || inv.Dying {
					continue
				}
				if rectsOverlap(b.X, b.Y, b.W, b.H, inv.X, inv.Y, inv.W, inv.H) {
					inv.Alive = false
					inv.Dying = true
					inv.DeathTick = 0
					g.spawnExplosion(inv.X+inv.W/2, inv.Y+inv.H/2)
					g.Score += 10
					g.InvaderSpeed += g.InvaderSpeedIncrease
					hit = true
					g.stopWheee(b.Wheee)
					b.Wheee = nil
					break
				}
			}
			if !hit {
				nextBullets = append(nextBullets, b)
			}
		}
		g.Bullets = nextBullets
	}

	// Win/Lose check (invaders only)
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
	// Speed up handled on each regular invader hit

	// Level complete when everything is finished
	if !g.LevelComplete && g.isLevelComplete() {
		g.LevelComplete = true
		g.CompleteTick = 0
		g.LevelState = levelStateCompleteWait
	}
	if g.LevelComplete {
		g.CompleteTick++
	}

	// Start transition on fire after completion text
	if g.LevelState == levelStateCompleteWait && g.CompleteTick >= completeTextDurationFrames {
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.LevelState = levelStateFadeWhite
			g.FadeTick = 0
		}
	}

	// Process kill-all sequence
	if g.KillAllActive {
		g.KillAllTick++
		if g.KillAllTick >= killAllIntervalFrames {
			g.KillAllTick = 0
			if len(g.KillAllOrder) > 0 {
				idx := g.KillAllOrder[0]
				g.KillAllOrder = g.KillAllOrder[1:]
				if idx >= 0 && idx < len(g.Invaders) {
					inv := &g.Invaders[idx]
					if inv.Alive {
						inv.Alive = false
						inv.Dying = true
						inv.DeathTick = 0
						g.spawnExplosion(inv.X+inv.W/2, inv.Y+inv.H/2)
						g.Score += 10
					}
				}
			}
			if len(g.KillAllOrder) == 0 {
				g.KillAllActive = false
			}
		}
	}

	g.updateEffects()

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

	if g.drawStartScreen(screen) {
		return
	}

	drawStarfield(screen, g.Stars)

	// Invaders
	for i := range g.Invaders {
		if !g.Invaders[i].Alive && !g.Invaders[i].Dying {
			continue
		}
		inv := &g.Invaders[i]
		if inv.Dying {
			t := float64(inv.DeathTick) / float64(deathAnimFrames-1)
			var scale float64
			if t < 0.3 {
				scale = 1.0 + (3.0-1.0)*(t/0.3)
			} else {
				scale = 3.0 * (1.0 - (t-0.3)/0.7)
			}
			if scale < 0 {
				scale = 0
			}
			alpha := 1.0 - t
			if alpha < 0 {
				alpha = 0
			}
			if inv.Type >= 0 && inv.Type < len(g.InvaderSprites) && g.Frame < len(g.InvaderSprites[inv.Type]) {
				op := &ebiten.DrawImageOptions{}
				cx := inv.W / 2
				cy := inv.H / 2
				op.GeoM.Translate(-cx, -cy)
				op.GeoM.Scale(scale, scale)
				op.GeoM.Translate(cx, cy)
				op.GeoM.Translate(inv.X, inv.Y)
				op.ColorM.Scale(1, 1, 1, alpha)
				screen.DrawImage(g.InvaderSprites[inv.Type][g.Frame], op)
			} else {
				ebitenutil.DrawRect(screen, inv.X, inv.Y, inv.W*scale, inv.H*scale, color.RGBA{R: 200, G: 70, B: 60, A: uint8(255 * alpha)})
			}
			continue
		}
		if inv.Type >= 0 && inv.Type < len(g.InvaderSprites) && g.Frame < len(g.InvaderSprites[inv.Type]) {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(inv.X, inv.Y)
			screen.DrawImage(g.InvaderSprites[inv.Type][g.Frame], op)
		} else {
			ebitenutil.DrawRect(screen, inv.X, inv.Y, inv.W, inv.H, color.RGBA{R: 200, G: 70, B: 60, A: 255})
		}
	}

	// UFO
	if g.UFO.Active && len(g.UFOSprites) == ufoFrameCount {
		op := &ebiten.DrawImageOptions{}
		if g.UFO.Crashing {
			cx := float64(ufoWidth) / 2
			cy := float64(ufoHeight) / 2
			op.GeoM.Translate(-cx, -cy)
			op.GeoM.Rotate(g.UFO.CrashAngle)
			op.GeoM.Translate(cx, cy)
		}
		op.GeoM.Translate(g.UFO.X, g.UFO.Y)
		screen.DrawImage(g.UFOSprites[g.UFO.Frame], op)
	}

	// C64 formation
	if g.Formation.Active && len(g.C64Sprites) == c64SpriteRows {
		t := float64(g.Formation.Tick) / float64(g.Formation.Duration)
		if t < 0 {
			t = 0
		}
		if t > 1 {
			t = 1
		}
		scale := 1.0
		if t < 0.45 {
			scale = 1.0 + 1.0*easeInOutQuad(t/0.45)
		} else if t < 0.7 {
			scale = 2.0
		} else {
			scale = 2.0 - 1.0*easeInOutQuad((t-0.7)/0.3)
		}
		for i := range g.Formation.Enemies {
			e := &g.Formation.Enemies[i]
			if e.SpriteIndex < 0 || e.SpriteIndex >= len(g.C64Sprites) {
				continue
			}
			op := &ebiten.DrawImageOptions{}
			cx := float64(c64SpriteWidth) / 2
			cy := float64(c64SpriteHeight) / 2
			op.GeoM.Translate(-cx, -cy)
			op.GeoM.Scale(scale, scale)
			op.GeoM.Rotate(e.Angle)
			op.GeoM.Translate(cx, cy)
			op.GeoM.Translate(e.X, e.Y)
			screen.DrawImage(g.C64Sprites[e.SpriteIndex][g.C64Frame], op)
		}
	}

	// Bullet
	for i := range g.Bullets {
		b := &g.Bullets[i]
		// Glow layers
		ebitenutil.DrawRect(screen, b.X-3, b.Y-4, b.W+6, b.H+8, color.RGBA{R: 120, G: 200, B: 255, A: 60})
		ebitenutil.DrawRect(screen, b.X-1, b.Y-2, b.W+2, b.H+4, color.RGBA{R: 180, G: 230, B: 255, A: 120})
		ebitenutil.DrawRect(screen, b.X, b.Y, b.W, b.H, color.RGBA{R: 240, G: 240, B: 255, A: 255})
	}

	// Player (draw last to stay on top)
	playerAlpha := 1.0
	if g.GameOver {
		t := float64(g.GameOverTick) / float64(gameOverShipFadeFrames)
		if t > 1 {
			t = 1
		}
		playerAlpha = 1 - t
	}
	playerImg := g.playerSpriteImage()
	if playerImg != nil {
		op := &ebiten.DrawImageOptions{}
		w := playerImg.Bounds().Dx()
		h := playerImg.Bounds().Dy()
		op.GeoM.Translate(g.Player.X-float64(w-int(g.Player.W))/2, g.Player.Y-float64(h-int(g.Player.H))/2)
		op.ColorM.Scale(1, 1, 1, playerAlpha)
		screen.DrawImage(playerImg, op)
	} else {
		ebitenutil.DrawRect(screen, g.Player.X, g.Player.Y, g.Player.W, g.Player.H, color.RGBA{R: 40, G: 200, B: 120, A: uint8(255 * playerAlpha)})
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

	drawHUD(screen, g)
	if g.GameOver {
		drawAmigaTextZoom(screen, g.StartFont, "GAME OVER", screenWidth/2, int(float64(screenHeight)*0.45), g.GameOverTick, gameOverZoomFrames, 4.0, 2.0)
	}

	// GREAT WORK banner
	if g.LevelComplete {
		t := float64(g.CompleteTick) / float64(completeTextDurationFrames)
		if t > 1 {
			t = 1
		}
		finalY := float64(screenHeight) * 0.42
		y := completeTextStartY + (finalY-completeTextStartY)*easeOutQuad(t)
		pulse := 1.0 + 0.08*math.Sin(float64(g.CompleteTick)*0.25)
		drawHugeText(screen, "GREAT WORK", screenWidth/2, int(y), pulse)
		if g.CompleteTick >= completeTextDurationFrames {
			glow := 0.5 + 0.5*math.Sin(float64(g.CompleteTick)*0.08)
			drawGlowText(screen, "PRESS FIRE TO CONTINUE", screenWidth/2, int(float64(screenHeight)*0.5), 2.2, glow)
		}
	}

	// Fade overlays
	if g.LevelState == levelStateFadeWhite {
		alpha := easeInOutQuad(float64(g.FadeTick) / float64(fadeWhiteFrames))
		if alpha > 1 {
			alpha = 1
		}
		drawFadeOverlay(screen, 1, 1, 1, alpha)
	}
	if g.LevelState == levelStateFadeBlack {
		alpha := float64(g.FadeTick) / float64(fadeBlackFrames)
		if alpha > 1 {
			alpha = 1
		}
		// Start from white, fade to black
		drawFadeOverlay(screen, 1, 1, 1, 1.0)
		drawFadeOverlay(screen, 0, 0, 0, alpha)
	}

	if g.GameOver && g.GameOverFading {
		alpha := float64(g.GameOverFadeTick) / float64(gameOverFadeFrames)
		if alpha > 1 {
			alpha = 1
		}
		drawFadeOverlay(screen, 0, 0, 0, alpha)
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
	g.playSfx(g.Sfx.Explosion, sfxVolume)
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
			X:       x,
			Y:       y,
			Vx:      vx,
			Vy:      vy,
			Life:    particleLifeMax,
			MaxLife: particleLifeMax,
			Color:   c,
			Size:    2 + g.Rand.Float64()*2,
			Gravity: 0.05,
		})
	}

	g.Shockwaves = append(g.Shockwaves, Shockwave{
		X:         x,
		Y:         y,
		Radius:    4,
		Life:      shockwaveLifeMax,
		MaxLife:   shockwaveLifeMax,
		Speed:     3.2,
		Thickness: 1.5,
		Color:     color.RGBA{R: 180, G: 220, B: 255, A: 255},
	})
}

func (g *Game) spawnBigExplosion(x, y float64) {
	g.playSfx(g.Sfx.Explosion, sfxVolume)
	colors := []color.RGBA{
		{R: 255, G: 240, B: 140, A: 255},
		{R: 255, G: 180, B: 80, A: 255},
		{R: 255, G: 120, B: 180, A: 255},
		{R: 140, G: 220, B: 255, A: 255},
		{R: 255, G: 90, B: 60, A: 255},
	}
	count := particleCount * 2
	for i := 0; i < count; i++ {
		angle := g.Rand.Float64() * 2 * math.Pi
		speed := 2.0 + g.Rand.Float64()*4.0
		vx := math.Cos(angle) * speed
		vy := math.Sin(angle) * speed
		c := colors[g.Rand.Intn(len(colors))]
		g.Particles = append(g.Particles, Particle{
			X:       x,
			Y:       y,
			Vx:      vx,
			Vy:      vy,
			Life:    particleLifeMax + 10,
			MaxLife: particleLifeMax + 10,
			Color:   c,
			Size:    3 + g.Rand.Float64()*3,
			Gravity: 0.04,
		})
	}

	g.Shockwaves = append(g.Shockwaves, Shockwave{
		X:         x,
		Y:         y,
		Radius:    6,
		Life:      shockwaveLifeMax + 8,
		MaxLife:   shockwaveLifeMax + 8,
		Speed:     4.2,
		Thickness: 2.2,
		Color:     color.RGBA{R: 220, G: 240, B: 255, A: 255},
	})
}

func (g *Game) spawnHugeExplosion(x, y float64) {
	g.playSfx(g.Sfx.Explosion, sfxVolume)
	colors := []color.RGBA{
		{R: 255, G: 250, B: 160, A: 255},
		{R: 255, G: 190, B: 90, A: 255},
		{R: 255, G: 140, B: 210, A: 255},
		{R: 160, G: 230, B: 255, A: 255},
		{R: 255, G: 100, B: 70, A: 255},
	}
	count := particleCount * hugeExplosionParticleMultiplier
	for i := 0; i < count; i++ {
		angle := g.Rand.Float64() * 2 * math.Pi
		speed := 2.6 + g.Rand.Float64()*5.2
		vx := math.Cos(angle) * speed
		vy := math.Sin(angle) * speed
		c := colors[g.Rand.Intn(len(colors))]
		g.Particles = append(g.Particles, Particle{
			X:       x,
			Y:       y,
			Vx:      vx,
			Vy:      vy,
			Life:    particleLifeMax + 16,
			MaxLife: particleLifeMax + 16,
			Color:   c,
			Size:    4 + g.Rand.Float64()*5,
			Gravity: 0.05,
		})
	}

	scale := hugeExplosionShockwaveScale
	life := int(float64(shockwaveLifeMax)*scale + 0.5)
	g.Shockwaves = append(g.Shockwaves, Shockwave{
		X:         x,
		Y:         y,
		Radius:    8,
		Life:      life,
		MaxLife:   life,
		Speed:     5.0,
		Thickness: 2.8,
		Color:     color.RGBA{R: 240, G: 250, B: 255, A: 255},
	})
}

func (g *Game) spawnGameOverExplosion(x, y float64) {
	g.playSfx(g.Sfx.Explosion, sfxVolume)
	colors := []color.RGBA{
		{R: 255, G: 250, B: 200, A: 255},
		{R: 255, G: 210, B: 120, A: 255},
		{R: 255, G: 160, B: 90, A: 255},
		{R: 200, G: 240, B: 255, A: 255},
		{R: 255, G: 120, B: 80, A: 255},
	}
	count := particleCount * hugeExplosionParticleMultiplier * 3
	for i := 0; i < count; i++ {
		angle := g.Rand.Float64() * 2 * math.Pi
		speed := 3.0 + g.Rand.Float64()*7.5
		vx := math.Cos(angle) * speed
		vy := math.Sin(angle) * speed
		c := colors[g.Rand.Intn(len(colors))]
		g.Particles = append(g.Particles, Particle{
			X:       x,
			Y:       y,
			Vx:      vx,
			Vy:      vy,
			Life:    particleLifeMax + 30,
			MaxLife: particleLifeMax + 30,
			Color:   c,
			Size:    5 + g.Rand.Float64()*7,
			Gravity: 0.05,
		})
	}

	for i := 0; i < 3; i++ {
		scale := 1.6 + float64(i)*0.3
		life := int(float64(shockwaveLifeMax)*scale + 0.5)
		g.Shockwaves = append(g.Shockwaves, Shockwave{
			X:         x,
			Y:         y,
			Radius:    10 + float64(i)*4,
			Life:      life,
			MaxLife:   life,
			Speed:     6.0 + float64(i)*1.2,
			Thickness: 3.2 + float64(i)*0.6,
			Color:     color.RGBA{R: 250, G: 255, B: 255, A: 255},
		})
	}

	for i := 0; i < 140; i++ {
		ox := (g.Rand.Float64() - 0.5) * 40
		oy := (g.Rand.Float64() - 0.5) * 30
		g.spawnSmokeBlack(x+ox, y+oy)
	}
}

func (g *Game) spawnSmoke(x, y float64) {
	c := []color.RGBA{
		{R: 120, G: 140, B: 160, A: 255},
		{R: 90, G: 110, B: 130, A: 255},
		{R: 160, G: 180, B: 200, A: 255},
	}[g.Rand.Intn(3)]
	vx := (g.Rand.Float64() - 0.5) * 0.6
	vy := 0.4 + g.Rand.Float64()*0.6
	g.Particles = append(g.Particles, Particle{
		X:       x,
		Y:       y,
		Vx:      vx,
		Vy:      vy,
		Life:    smokeLifeMax,
		MaxLife: smokeLifeMax,
		Color:   c,
		Size:    2 + g.Rand.Float64()*2,
		Gravity: 0.02,
	})
}

func (g *Game) spawnSmokeBlack(x, y float64) {
	c := []color.RGBA{
		{R: 30, G: 30, B: 30, A: 255},
		{R: 20, G: 20, B: 20, A: 255},
		{R: 50, G: 50, B: 50, A: 255},
	}[g.Rand.Intn(3)]
	vx := (g.Rand.Float64() - 0.5) * 0.4
	vy := 0.6 + g.Rand.Float64()*0.8
	g.Particles = append(g.Particles, Particle{
		X:       x + (g.Rand.Float64()-0.5)*4,
		Y:       y + (g.Rand.Float64()-0.5)*4,
		Vx:      vx,
		Vy:      vy,
		Life:    smokeLifeMax + 6,
		MaxLife: smokeLifeMax + 6,
		Color:   c,
		Size:    4 + g.Rand.Float64()*6,
		Gravity: 0.03,
	})
}

func (g *Game) resetToStartScreen() {
	g.GameOver = false
	g.GameOverTriggered = false
	g.GameOverTick = 0
	g.GameOverWaitTick = 0
	g.GameOverFadeTick = 0
	g.GameOverFading = false

	g.Level = 1
	g.Score = 0
	g.Lives = 3
	g.LevelComplete = false
	g.CompleteTick = 0
	g.LevelState = levelStatePlaying
	g.FadeTick = 0

	g.Bullets = g.Bullets[:0]
	g.Particles = g.Particles[:0]
	g.Shockwaves = g.Shockwaves[:0]
	g.UFO = UFO{}
	g.UFOSpawnTick = ufoSpawnFrames
	g.Formation = Formation{Active: false, Tick: 0, Duration: formationDurationFrames}
	g.FormationSpawnTick = formationInitialDelayFrames

	g.StartState = startStateScreen
	g.StartTick = 0
	g.StartFadeTick = 0
	g.startMusic("music/codex_amiga_mstb.mod")
	g.setMusicVolume(0.3)

	g.resetLevelState()
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

func (g *Game) startFormation() {
	g.Formation.Active = true
	g.Formation.Tick = 0
	g.Formation.Duration = formationDurationFrames
	if g.Formation.Side == 0 {
		g.Formation.Side = 1
	} else {
		g.Formation.Side = 0
	}
	for i := 0; i < len(g.Formation.Enemies); i++ {
		row := i / 4
		col := i % 4
		side := g.Formation.Side
		x, y := formationPos(side, row, col, 0)
		spriteIndex := g.Rand.Intn(c64SpriteRows)
		g.Formation.Enemies[i] = FlyEnemy{
			Index:       i,
			Side:        side,
			SpriteIndex: spriteIndex,
			X:           x,
			Y:           y,
			PrevX:       x,
			PrevY:       y,
			Angle:       0,
		}
	}
}

func (g *Game) isLevelComplete() bool {
	if !g.Win {
		return false
	}
	for i := range g.Invaders {
		if g.Invaders[i].Alive || g.Invaders[i].Dying {
			return false
		}
	}
	return true
}

func (g *Game) resetLevelState() {
	g.LevelComplete = false
	g.CompleteTick = 0
	g.LevelState = levelStatePlaying
	g.FadeTick = 0
	g.Win = false

	for i := range g.Bullets {
		g.stopWheee(g.Bullets[i].Wheee)
	}
	g.Bullets = g.Bullets[:0]
	g.Particles = g.Particles[:0]
	g.Shockwaves = g.Shockwaves[:0]
	g.KillAllActive = false
	g.KillAllTick = 0
	g.KillAllOrder = g.KillAllOrder[:0]

	g.UFO = UFO{}
	g.UFOSpawnTick = ufoSpawnFrames
	g.stopUFOLoop()
	g.stopSuperLoop()

	g.Formation.Active = false
	g.Formation.Tick = 0
	g.Formation.Duration = formationDurationFrames
	g.FormationSpawnTick = formationInitialDelayFrames

	g.Invaders = make([]Invader, 0, invaderCols*invaderRows)
	for row := 0; row < invaderRows; row++ {
		for col := 0; col < invaderCols; col++ {
			x := float64(invaderStartX + col*(invaderWidth+invaderGapX))
			y := float64(invaderStartY + row*(invaderHeight+invaderGapY))
			g.Invaders = append(g.Invaders, Invader{
				X:     x,
				Y:     y,
				W:     invaderWidth,
				H:     invaderHeight,
				Alive: true,
				Type:  row,
			})
		}
	}
	g.InvaderDir = 1
	g.InvaderSpeed = 1.0
	g.InvaderSpeedThreshold = (invaderCols * invaderRows) / 2
	g.InvaderSpeedIncrease = float64(g.Level) / 5.0
}

func (g *Game) updateEffects() {
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

	// Invader death animations
	for i := range g.Invaders {
		if g.Invaders[i].Dying {
			g.Invaders[i].DeathTick++
			if g.Invaders[i].DeathTick >= deathAnimFrames {
				g.Invaders[i].Dying = false
			}
		}
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
}
