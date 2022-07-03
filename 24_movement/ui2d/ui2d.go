package ui2d

import (
	"bufio"
	"fmt"
	"image/png"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/pdelewski/games-with-go/24_movement/game"
	"github.com/veandco/go-sdl2/sdl"
)

const winWidth, winHeight int = 1280, 720

var renderer *sdl.Renderer
var textureAtlas *sdl.Texture // Spritesheets called texture atlases

var textureIndex map[game.Tile][]sdl.Rect // Go map from a tile to rect

var prevKeyboardState []uint8
var keyboardState []uint8

func loadTextureIndex() {
	textureIndex = make(map[game.Tile][]sdl.Rect)
	infile, err := os.Open("ui2d/assets/atlas-index.txt")
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	// Read from scanner
	scanner := bufio.NewScanner(infile) // *File satisfies io.Reader interface
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line) // Remove extra spaces
		tileRune := game.Tile(line[0]) // Get first rune from the string
		xy := line[1:]                 // Get ButFirst
		splitXYC := strings.Split(xy, ",")
		x, err := strconv.ParseInt(strings.TrimSpace(splitXYC[0]), 10, 64) // base10, bit size 64
		if err != nil {
			panic(err)
		}
		y, err := strconv.ParseInt(strings.TrimSpace(splitXYC[1]), 10, 64)
		if err != nil {
			panic(err)
		}
		// Tile variation
		variationCount, err := strconv.ParseInt(strings.TrimSpace(splitXYC[2]), 10, 64)
		if err != nil {
			panic(err)
		}
		var rects []sdl.Rect
		for i := int64(0); i < variationCount; i++ {
			rects = append(rects, sdl.Rect{int32(x * 32), int32(y * 32), 32, 32})
			// Wrap around if varied images continue on a new line
			x++
			if x > 62 {
				x = 0
				y++
			}
		}
		textureIndex[tileRune] = rects
	}
}

func imgFileToTexture(filename string) *sdl.Texture {
	// Open
	infile, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	// Decode
	img, err := png.Decode(infile)
	if err != nil {
		panic(err)
	}

	// Extract w/h
	w := img.Bounds().Max.X
	h := img.Bounds().Max.Y

	pixels := make([]byte, w*h*4)
	bIndex := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixels[bIndex] = byte(r / 256)
			bIndex++
			pixels[bIndex] = byte(g / 256)
			bIndex++
			pixels[bIndex] = byte(b / 256)
			bIndex++
			pixels[bIndex] = byte(a / 256)
			bIndex++
		}
	}

	// Make an SDL2 texture out of pixels
	// AGBR is backwards from way we will be filling in out bytes
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	tex.Update(nil, pixels, w*4) // Can't provide a rectangle, pitch = 4 bytes per pixel

	// Set blend mode to alpha blending
	err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		panic(err)
	}
	return tex
}

// Init callback runs before anything else
func init() {
	// Check what best performing, most reliable renderer SDL is using
	// (direct3d on windows, opengl on linux, or software for just sdl)
	sdl.LogSetAllPriority(sdl.LOG_PRIORITY_VERBOSE)

	// Initialize SDL2.
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create a window.
	window, err := sdl.CreateWindow("RPG", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(winWidth), int32(winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create renderer.
	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Set hints.
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	// Create texture.
	textureAtlas = imgFileToTexture("../22_texture_index/ui2d/assets/tiles.png")

	loadTextureIndex()

	// Update keyboard state
	keyboardState = sdl.GetKeyboardState() // Updates by sdl
	prevKeyboardState = make([]uint8, len(keyboardState))
	for i, v := range keyboardState {
		prevKeyboardState[i] = v
	}
}

// UI2d ...
type UI2d struct {
}

// Draw generates a random (but reproducable) tile variety
func (ui *UI2d) Draw(level *game.Level) {
	// Clear before drawing tiles
	renderer.Clear()

	rand.Seed(1)
	for y, row := range level.Map {
		for x, tile := range row {
			if tile != game.Blank {
				srcRects := textureIndex[tile]
				srcRect := srcRects[rand.Intn(len(srcRects))] // Random number between 1 and length of variations
				dstRect := sdl.Rect{int32(x * 32), int32(y * 32), 32, 32}
				renderer.Copy(textureAtlas, &srcRect, &dstRect)
			}
		}
	}
	// Draw player sprite (21,59) ontop of tiles
	renderer.Copy(textureAtlas, &sdl.Rect{21 * 32, 59 * 32, 32, 32}, &sdl.Rect{int32(level.Player.X) * 32, int32(level.Player.Y) * 32, 32, 32})
	// renderer.Copy(textureAtlas, nil, nil) // Draw whole texture atlas
	renderer.Present()
}

// GetInput polls for events, and quits when event is nil
func (ui *UI2d) GetInput() *game.Input {
	// Keep waiting for user input
	for {
		// Poll for events
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return &game.Input{Typ: game.Quit}
			}
		}

		// Handle keypresses
		var input game.Input
		if keyboardState[sdl.SCANCODE_UP] != 0 && prevKeyboardState[sdl.SCANCODE_UP] == 0 {
			input.Typ = game.Up
		}
		if keyboardState[sdl.SCANCODE_DOWN] != 0 && prevKeyboardState[sdl.SCANCODE_DOWN] == 0 {
			input.Typ = game.Down
		}
		if keyboardState[sdl.SCANCODE_LEFT] != 0 && prevKeyboardState[sdl.SCANCODE_LEFT] == 0 {
			input.Typ = game.Left
		}
		if keyboardState[sdl.SCANCODE_RIGHT] != 0 && prevKeyboardState[sdl.SCANCODE_RIGHT] == 0 {
			input.Typ = game.Right
		}

		// Update previous keyboard state
		for i, v := range keyboardState {
			prevKeyboardState[i] = v
		}

		if input.Typ != game.None {
			return &input
		}
	}
}
