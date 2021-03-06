package main

// To Do
// [ ]	Some bug with the last few balloons. Can't click and they keep speeding up.
// [ ]	Try using insertion sort for balloons  - https://en.wikipedia.org/wiki/Insertion_sort
// [ ]	Time it to see if you can beat Go's built sort
// [ ]	Implement collisions with the balloons
//			- Treat balloons as spheres - only check distance between center of balloons (use Vec3)
//			- Minimum translation vector (see pong.go) [hard]

import (
	"fmt"
	"image/png"
	"math"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/stephen-mahon/games-with-go/noise"
	. "github.com/stephen-mahon/games-with-go/vec3"
	"github.com/veandco/go-sdl2/sdl"
)

const winWidth, winHeight, winDepth int = 1280, 720, 100

type audioState struct {
	explosionBytes []byte
	deviceID       sdl.AudioDeviceID
	audioSpec      *sdl.AudioSpec
}

type mouseState struct {
	leftButton, rightButton bool
	x, y                    int
}

func getMouseState() mouseState {
	mouseX, mouseY, mouseButtonState := sdl.GetMouseState()
	leftButton := mouseButtonState & sdl.ButtonLMask()
	rightButton := mouseButtonState & sdl.ButtonRMask()
	var result mouseState
	result.x = int(mouseX)
	result.y = int(mouseY)
	result.leftButton = !(leftButton == 0)
	result.rightButton = !(rightButton == 4)
	return result
}

type balloon struct {
	tex  *sdl.Texture
	pos  Vector3
	dir  Vector3
	w, h int

	exploding, exploded bool
	explosionStart      time.Time
	explosionInterval   float32
	explosionTexture    *sdl.Texture
}

func newBalloon(tex *sdl.Texture, pos, dir Vector3, explosionTexture *sdl.Texture) *balloon {
	_, _, w, h, err := tex.Query()
	if err != nil {
		panic(err)
	}
	return &balloon{tex, pos, dir, int(w), int(h), false, false, time.Now(), 20, explosionTexture}
}

type balloonArray []*balloon

func (balloons balloonArray) Len() int {
	return len(balloons)
}

func (balloons balloonArray) Swap(i, j int) {
	balloons[i], balloons[j] = balloons[j], balloons[i]
}

func (balloons balloonArray) Less(i, j int) bool {
	diff := balloons[i].pos.Z - balloons[j].pos.Z
	return diff < -1
}

func (balloon *balloon) getScale() float32 {
	return (balloon.pos.Z/200 + 1) / 2
}

func (balloon *balloon) getCircle() (x, y, r float32) {
	x = balloon.pos.X
	y = balloon.pos.Y - 30*balloon.getScale()
	r = float32(balloon.w) * balloon.getScale() / 2
	return x, y, r
}

func updateBalloons(balloons []*balloon, elaspedTime float32,
	currentMouseState mouseState,
	prevMouseState mouseState,
	audioState *audioState) []*balloon {

	numAnimation := 16
	balloonClicked := false
	balloonsExploded := false
	for i := len(balloons) - 1; i >= 0; i-- {
		balloon := balloons[i]

		if balloon.exploding {
			animationElasped := float32(time.Since(balloon.explosionStart).Seconds() * 1000)
			animationIndex := numAnimation - 1 - int(animationElasped/balloon.explosionInterval)
			if animationIndex < 0 {
				balloon.exploding = false
				balloon.exploded = true
				balloonsExploded = true
			}
		}

		if !balloonClicked && !prevMouseState.leftButton && currentMouseState.leftButton {
			x, y, r := balloon.getCircle()
			mouseX := currentMouseState.x
			mouseY := currentMouseState.y
			xDiff := float32(mouseX) - x
			yDiff := float32(mouseY) - y
			dist := float32(math.Sqrt(float64(xDiff*xDiff + yDiff*yDiff)))
			if dist < r {
				balloonClicked = true
				sdl.ClearQueuedAudio(audioState.deviceID)
				sdl.QueueAudio(audioState.deviceID, audioState.explosionBytes)
				sdl.PauseAudioDevice(audioState.deviceID, false)
				balloon.exploding = true
				balloon.explosionStart = time.Now()
			}
		}
		p := Add(balloon.pos, Mult(balloon.dir, elaspedTime))

		if p.X < 0 || p.X > float32(winWidth) {
			balloon.dir.X = -balloon.dir.X
		}
		if p.Y < 0 || p.Y > float32(winHeight) {
			balloon.dir.Y = -balloon.dir.Y
		}
		if p.Z < 0 || p.Z > float32(winDepth) {
			balloon.dir.Z = -balloon.dir.Z
		}

		balloon.pos = Add(balloon.pos, Mult(balloon.dir, elaspedTime))
	}

	if balloonsExploded {
		filteredBalloons := balloons[0:0]
		for _, balloon := range balloons {
			if !balloon.exploded {
				filteredBalloons = append(filteredBalloons, balloon)
			}
		}
		balloons = filteredBalloons
	}
	return balloons
}

func (balloon *balloon) draw(renderer *sdl.Renderer) {
	scale := balloon.getScale()
	newW := int32(float32(balloon.w) * scale)
	newH := int32(float32(balloon.h) * scale)
	x := int32(balloon.pos.X - float32(newW)/2)
	y := int32(balloon.pos.Y - float32(newH)/2)
	rect := &sdl.Rect{x, y, newW, newH}
	renderer.Copy(balloon.tex, nil, rect)

	if balloon.exploding {
		numAnimation := 16
		animationElasped := float32(time.Since(balloon.explosionStart).Seconds() * 1000)
		animationIndex := numAnimation - 1 - int(animationElasped/balloon.explosionInterval)
		animationX := animationIndex % 4
		animationY := 64 * ((animationIndex - animationX) / 4)
		animationX *= 64
		animationRect := &sdl.Rect{int32(animationX), int32(animationY), 64, 64}
		rect.X -= rect.W / 2
		rect.Y -= rect.H / 2
		rect.W *= 2
		rect.H *= 2
		renderer.Copy(balloon.explosionTexture, animationRect, rect)
	}
}

type rgba struct {
	r, g, b byte
}

func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}

func setPixel(x, y int, c rgba, pixels []byte) {
	index := (y*winWidth + x) * 4
	if index < len(pixels)-4 && index >= 0 {
		pixels[index] = c.r
		pixels[index+1] = c.g
		pixels[index+2] = c.b
	}
}

func pixelsToTexture(renderer *sdl.Renderer, pixels []byte, w, h int) *sdl.Texture {
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	tex.Update(nil, pixels, w*4)
	return tex
}

func imgFileToTexture(renderer *sdl.Renderer, filename string) *sdl.Texture {

	infile, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer infile.Close()
	// png decoder outputs an image with a premultiplied alpha rgb, i.e. ((r*alpha),(g*alpha),(b*alpha),alpha)
	img, err := png.Decode(infile)
	if err != nil {
		panic(err)
	}

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
	tex := pixelsToTexture(renderer, pixels, w, h)
	err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		panic(err)
	}
	return tex
}

func loadBalloon(renderer *sdl.Renderer, numBallons int) []*balloon {

	explosionTexture := imgFileToTexture(renderer, "explosion.png")

	balloonStrs := []string{"balloon_red.png", "balloon_green.png", "balloon_blue.png"}
	balloonTextures := make([]*sdl.Texture, len(balloonStrs))

	for i, bstr := range balloonStrs {
		balloonTextures[i] = imgFileToTexture(renderer, bstr)
	}

	balloons := make([]*balloon, numBallons)
	for i := range balloons {
		tex := balloonTextures[i%3]
		pos := Vector3{rand.Float32() * float32(winWidth), rand.Float32() * float32(winHeight), rand.Float32() * float32(winDepth)}
		dir := Vector3{rand.Float32()*0.5 - 0.25, rand.Float32()*0.5 - 0.25, rand.Float32()*0.25 - 0.25/2}
		balloons[i] = newBalloon(tex, pos, dir, explosionTexture)
	}

	return balloons
}

func lerp(b1 byte, b2 byte, pct float32) byte {
	return byte(float32(b1) + pct*(float32(b2)-float32(b1)))
}

func colorLerp(c1, c2 rgba, pct float32) rgba {
	return rgba{lerp(c1.r, c2.r, pct), lerp(c1.g, c2.g, pct), lerp(c1.b, c2.b, pct)}
}

func getGradient(c1, c2 rgba) []rgba {
	result := make([]rgba, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		result[i] = colorLerp(c1, c2, pct)
	}
	return result
}

func getDualGradient(c1, c2, c3, c4 rgba) []rgba {
	result := make([]rgba, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		if pct < 0.5 {
			result[i] = colorLerp(c1, c2, pct*float32(2))
		} else {
			result[i] = colorLerp(c3, c4, pct*float32(1.5)-float32(0.5))
		}
	}
	return result
}

func clamp(min, max, v int) int {
	if v < min {
		v = min
	} else if v > max {
		v = max
	}
	return v
}

func rescaleAndDraw(noise []float32, min, max float32, gradient []rgba, w, h int) []byte {
	result := make([]byte, w*h*4)
	scale := 255.0 / (max - min)
	offset := min * scale

	for i := range noise {
		noise[i] = noise[i]*scale - offset
		c := gradient[clamp(0, 255, int(noise[i]))]
		p := i * 4
		result[p] = c.r
		result[p+1] = c.g
		result[p+2] = c.b
	}
	return result
}

func main() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Exploding Balloons", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(winWidth), int32(winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer renderer.Destroy()

	explosionBytes, audioSpec := sdl.LoadWAV("explode.wav")
	audioID, err := sdl.OpenAudioDevice("", false, audioSpec, nil, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sdl.FreeWAV(explosionBytes)

	audioState := audioState{explosionBytes, audioID, audioSpec}

	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	cloudNoise, min, max := noise.MakeNoise(noise.FBM, .009, .5, 3, 3, winWidth, winHeight)
	cloudGradient := getGradient(rgba{0, 0, 255}, rgba{255, 255, 255})
	cloudPixels := rescaleAndDraw(cloudNoise, min, max, cloudGradient, winWidth, winHeight)
	cloudTexture := pixelsToTexture(renderer, cloudPixels, winWidth, winHeight)

	balloons := loadBalloon(renderer, 3)
	var elaspedTime float32
	currentMouseState := getMouseState()
	prevMouseState := currentMouseState

	for {
		frameStart := time.Now()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}
		currentMouseState = getMouseState()

		renderer.Copy(cloudTexture, nil, nil)
		balloons := updateBalloons(balloons, elaspedTime, currentMouseState, prevMouseState, &audioState)
		sort.Stable(balloonArray(balloons))
		for _, balloon := range balloons {
			balloon.draw(renderer)
		}

		renderer.Present()
		// basic framerate independence setup
		elaspedTime = float32(time.Since(frameStart).Seconds() * 1000)
		//fmt.Println("ms per frame:", elaspedTime)
		if elaspedTime < 5 {
			sdl.Delay(5 - uint32(elaspedTime))
			elaspedTime = float32(time.Since(frameStart).Seconds() * 1000)
		}
		prevMouseState = currentMouseState

	}

}
