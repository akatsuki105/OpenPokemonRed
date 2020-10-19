package game

import (
	"pokered/pkg/audio"
	"pokered/pkg/data/tileset"
	"pokered/pkg/data/worldmap"
	"pokered/pkg/data/worldmap/header"
	"pokered/pkg/joypad"
	"pokered/pkg/script"
	"pokered/pkg/sprite"
	"pokered/pkg/store"
	"pokered/pkg/util"
	"pokered/pkg/world"
)

func execOverworld() {
	p := store.SpriteData[0]
	if p == nil {
		return
	}

	if util.ReadBit(store.D736, 6) {
		sprite.HandleMidJump()
	}

	if p.WalkCounter > 0 {
		sprite.UpdateSprites()
		sprite.AdvancePlayerSprite()
	} else {
		joypadOverworld()

		directionPressed := false
		switch {
		case joypad.JoyHeld.Start:
			audio.PlaySound(audio.SFX_START_MENU)
			script.SetID(script.WidgetStartMenu)
			return
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
	if util.ReadBit(store.D736, 0) {
		util.ResBit(&store.D736, 0)
		sprite.StepOutFromDoor()
	}
}

func moveAhead() {
	checkWarpsNoCollision()
}

// check if the player has stepped onto a warp after having not collided
// ref: CheckWarpsNoCollision
func checkWarpsNoCollision() {
	curWorld := world.CurWorld
	if len(curWorld.Object.Warps) == 0 {
		checkMapConnections()
		return
	}

	p := store.SpriteData[0]
	if p == nil {
		return
	}
	for i, w := range curWorld.Object.Warps {
		if p.MapXCoord == w.XCoord && p.MapYCoord == w.YCoord {
			util.SetBit(&store.D736, 2)
			if sprite.IsPlayerStandingOnDoorOrWarp() {
				warpFound(i)
				return
			}

			if !extraWarpCheck() {
				return
			}

			// if the extra check passed
			joypad.Joypad()
			if joypad.JoyHeld.Down || joypad.JoyHeld.Up || joypad.JoyHeld.Left || joypad.JoyHeld.Right {
				warpFound(i)
			}

		}
	}

	checkMapConnections()
}

// ref: ExtraWarpCheck
func extraWarpCheck() bool {
	result := false
	curMap, curTileset := world.CurWorld.MapID, world.CurWorld.Header.Tileset

	switch curMap {
	case worldmap.ROCKET_HIDEOUT_B1F, worldmap.ROCKET_HIDEOUT_B2F, worldmap.ROCKET_HIDEOUT_B4F, worldmap.ROCK_TUNNEL_1F:
		result = sprite.IsWarpTileInFrontOfPlayer()

	default:
		switch curTileset {
		case tileset.Overworld, tileset.Ship, tileset.ShipPort, tileset.Plateau:
			result = sprite.IsWarpTileInFrontOfPlayer()
		default:
			result = sprite.IsPlayerFacingEdgeOfMap()
		}
	}
	return result
}

// ref: CheckMapConnections
func checkMapConnections() {
	curWorld := world.CurWorld
	p := store.SpriteData[0]
	if p == nil {
		return
	}

	if p.Direction == util.Up && p.MapYCoord == -1 {
		for i, XCoord := range curWorld.Header.Connections.North.Coords {
			if p.MapXCoord == int(XCoord) {
				destMapID := curWorld.Header.Connections.North.DestMapID
				DestMapHeader := header.Get(destMapID)
				loadWorldData(destMapID, -1)
				p.MapXCoord = int(DestMapHeader.Connections.South.Coords[i])
				p.MapYCoord = int(DestMapHeader.Height*2 - 1)
				return
			}
		}
	}

	if p.Direction == util.Down && p.MapYCoord == int(curWorld.Header.Height*2) {
		for i, XCoord := range curWorld.Header.Connections.South.Coords {
			if p.MapXCoord == int(XCoord) {
				destMapID := curWorld.Header.Connections.South.DestMapID
				DestMapHeader := header.Get(destMapID)
				loadWorldData(destMapID, -1)
				p.MapXCoord = int(DestMapHeader.Connections.North.Coords[i])
				p.MapYCoord = 0
				return
			}
		}
	}
}

func warpFound(warpID int) {
	if checkIfInOutsideMap() {
		world.LastWorld = world.CurWorld
		w := world.CurWorld.Object.Warps[warpID]
		destMapID := w.DestMap
		if destMapID != worldmap.ROCK_TUNNEL_1F {
		}
		playMapChangeSound()
		loadWorldData(destMapID, warpID)
		return
	}

	// indoorMaps
	destMapID := world.CurWorld.Object.Warps[warpID].DestMap
	if destMapID == worldmap.LAST_MAP {
		destMapID = world.LastWorld.MapID
	}
	playMapChangeSound()
	util.SetBit(&store.D736, 0)
	loadWorldData(destMapID, warpID)
}

// If the player is in an outside map (a town or route), set the z flag
func checkIfInOutsideMap() bool {
	tilesetID := world.CurBlockset.TilesetID
	return tilesetID == tileset.Overworld || tilesetID == tileset.Plateau
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
	warpTo := world.CurWorld.Object.WarpTos[warpID]

	// ref: LoadDestinationWarpPosition
	if warpID >= 0 {
		p := store.SpriteData[0]
		p.MapXCoord, p.MapYCoord = warpTo.XCoord, warpTo.YCoord
	}
}
