package preview

import (
	"image"
	"image/color"
	"math"
)

func clampUint8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func sampleSceneColor(sharp bool, w, h int, px, py int, sec float64) (uint8, uint8, uint8) {
	fy := float64(py) / float64(h)

	horizon := 0.70

	// Water / Lake at bottom
	if fy >= horizon {
		waterDepth := (fy - horizon) / (1.0 - horizon)
		baseR := 15.0 + waterDepth*12.0
		baseG := 23.0 + waterDepth*14.0
		baseB := 42.0 + waterDepth*18.0

		wave := math.Sin(float64(px)*0.08 + sec*3.0 + float64(py)*0.5)
		if wave > 0.6 {
			factor := (wave - 0.6) / 0.4
			baseR += factor * 80.0
			baseG += factor * 50.0
			baseB += factor * 30.0
		}
		return clampUint8(baseR), clampUint8(baseG), clampUint8(baseB)
	}

	// Mountain 2 (Near range silhouette)
	m2Y := 0.58 + 0.08*math.Sin(float64(px)*0.015+1.5) + 0.04*math.Cos(float64(px)*0.03)
	if fy >= m2Y {
		ridgeDiff := math.Abs(fy - m2Y)
		if sharp && ridgeDiff < 0.008 {
			return 70, 75, 95
		}
		return 30, 35, 48
	}

	// Mountain 1 (Distant range silhouette)
	m1Y := 0.48 + 0.07*math.Sin(float64(px)*0.02) + 0.03*math.Cos(float64(px)*0.045)
	if fy >= m1Y {
		return 55, 45, 75
	}

	// Sky Gradient (Deep twilight to sunset amber)
	t := fy / horizon
	var r, g, b float64
	if t < 0.45 {
		st := t / 0.45
		r = 15.0 + st*45.0
		g = 18.0 + st*25.0
		b = 40.0 + st*60.0
	} else if t < 0.8 {
		st := (t - 0.45) / 0.35
		r = 60.0 + st*150.0
		g = 43.0 + st*80.0
		b = 100.0 - st*50.0
	} else {
		st := (t - 0.8) / 0.2
		r = 210.0 + st*35.0
		g = 123.0 + st*45.0
		b = 50.0 - st*20.0
	}

	// Sun / Glowing celestial orb
	sunX := float64(w) * (0.65 + 0.05*math.Sin(sec*0.2))
	sunY := float64(h) * 0.32
	dx := float64(px) - sunX
	dy := float64(py) - sunY
	dist := math.Sqrt(dx*dx + dy*dy)
	sunRadius := 26.0

	if dist < sunRadius {
		if sharp {
			innerFactor := dist / sunRadius
			return clampUint8(255 - innerFactor*10), clampUint8(245 - innerFactor*30), clampUint8(210 - innerFactor*60)
		}
		return 255, 230, 180
	} else if dist < sunRadius*3.0 {
		corona := 1.0 - (dist - sunRadius) / (sunRadius * 2.0)
		r += corona * 100.0
		g += corona * 60.0
		b += corona * 20.0
	}

	// Twinkling stars
	if fy < 0.40 {
		starHash := int(float64(px)*37.0 + float64(py)*17.0)
		if starHash%149 == 0 {
			if sharp {
				r = math.Min(255, r+180)
				g = math.Min(255, g+180)
				b = math.Min(255, b+220)
			} else {
				r = math.Min(255, r+70)
				g = math.Min(255, g+70)
				b = math.Min(255, b+90)
			}
		}
	}

	return clampUint8(r), clampUint8(g), clampUint8(b)
}

// GenerateScenicSample creates a procedural test scenic image with procedural sharpness.
func GenerateScenicSample(sharp bool, w, h int, sec float64) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var r, g, b uint8
			if sharp {
				r, g, b = sampleSceneColor(true, w, h, x, y, sec)
			} else {
				bx := (x / 6) * 6
				by := (y / 7) * 7
				r, g, b = sampleSceneColor(false, w, h, bx, by, sec)
			}
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}
