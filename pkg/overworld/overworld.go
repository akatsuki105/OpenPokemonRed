package overworld

import (
	"pokered/pkg/audio"
	"pokered/pkg/joypad"
	"pokered/pkg/palette"
	"pokered/pkg/script"
	"pokered/pkg/sprite"
	"pokered/pkg/store"
	"pokered/pkg/text"
	"pokered/pkg/util"
	"pokered/pkg/world"
)

// ExecOverworld exec overworld loop
// ref: OverworldLoop
func ExecOverworld() {
	p := store.SpriteData[0]
	if p == nil {
		return
	}

	palette.LoadGBPal()

	if util.ReadBit(store.D736, 6) {
		sprite.HandleMidJump()
	}

	if p.WalkCounter > 0 {
		sprite.UpdateSprites()
		sprite.AdvancePlayerSprite()

		if p.WalkCounter == 0 {
			if (p.DeltaX + p.DeltaY) != 0 {
				store.Enable.NormalWarp = true
			}
		}
	} else {
		joypadOverworld()

		directionPressed := false
		switch {
		case joypad.JoyHeld.Start:
			audio.PlaySound(audio.SFX_START_MENU)
			script.SetID(script.WidgetStartMenu)
			return
		case joypad.JoyHeld.A:
			if spriteOffset := sprite.GetFrontSpriteOrSign(0); spriteOffset > 0 {
				displayDialogue(spriteOffset)
				return
			}
		case joypad.JoyHeld.Down:
			p.DeltaY = 1
			p.Direction = util.Down
		case joypad.JoyHeld.Up:
			p.DeltaY = -1
			p.Direction = util.Up
		case joypad.JoyHeld.Right:
			p.DeltaX = 1
			p.Direction = util.Right
		case joypad.JoyHeld.Left:
			p.DeltaX = -1
			p.Direction = util.Left
		}

		h := joypad.JoyHeld
		directionPressed = h.Up || h.Down || h.Right || h.Left
		if directionPressed {
			p.WalkCounter = 16
			sprite.UpdateSprites()
			if sprite.CollisionCheckForPlayer() {
				p.DeltaX, p.DeltaY = 0, 0
			}
			sprite.AdvancePlayerSprite()
		} else {
			sprite.UpdateSprites()
			p.RightHand = false
			return
		}
	}
	moveAhead()
}

// simulatedJoypad
func joypadOverworld() {
	p := store.SpriteData[0]
	p.DeltaX, p.DeltaY = 0, 0

	runMapScript()

	joypad.Joypad()

	if p.Direction == util.Down && sprite.IsStandingOnDoor(0) {
		joypad.JoyHeld = joypad.Input{Down: true}
		return
	}

	if len(p.Simulated) == 0 {
		return
	}

	switch p.Simulated[0] {
	case util.Down:
		joypad.JoyHeld = joypad.Input{Down: true}
	case util.Up:
		joypad.JoyHeld = joypad.Input{Up: true}
	case util.Right:
		joypad.JoyHeld = joypad.Input{Right: true}
	case util.Left:
		joypad.JoyHeld = joypad.Input{Left: true}
	}
	if len(p.Simulated) > 1 {
		p.Simulated = p.Simulated[1:]
		return
	}
	p.Simulated = []uint{}
}

// ref: RunMapScript
func runMapScript() {
	runNPCMovementScript()
}

// ref: RunNPCMovementScript
func runNPCMovementScript() {
}

func moveAhead() {
	checkWarpsNoCollision()
}

// function to play a sound when changing maps
func playMapChangeSound() {
	_, tileID := world.GetTileID(8, 8)
	soundID := audio.SFX_GO_OUTSIDE
	if tileID == 0x0b {
		soundID = audio.SFX_GO_INSIDE
	}
	audio.PlaySound(soundID)
}

func loadWorldData(mapID, warpID int) {
	world.LoadWorldData(mapID)

	// ref: LoadDestinationWarpPosition
	if warpID >= 0 {
		warpTo := world.CurWorld.Object.WarpTos[warpID]
		p := store.SpriteData[0]
		p.MapXCoord, p.MapYCoord = warpTo.XCoord, warpTo.YCoord
	}
}

func displayDialogue(offset int) {
	texts, textID := world.CurWorld.Header.Text, offset
	text.DisplayTextID(text.Image, texts, textID)
	script.SetID(script.ExecText)
}