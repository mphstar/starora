package engine

import (
	"fmt"
	"math"
	"runtime"
	"strings"
)

// ResolveModelName maps the user model selection string to the Real-ESRGAN model filename.
func ResolveModelName(baseModel string, scale int) string {
	if strings.Contains(baseModel, "animevideov3") {
		return fmt.Sprintf("realesr-animevideov3-x%d", scale)
	}
	if strings.HasPrefix(baseModel, "realesrgan-x4plus-anime") {
		return "realesrgan-x4plus-anime"
	}
	if strings.HasPrefix(baseModel, "realesrnet-x4plus") {
		return "realesrnet-x4plus"
	}
	return "realesrgan-x4plus"
}

// ParseTileSize extracts the tile-size integer parameter.
func ParseTileSize(s string) string {
	fields := strings.Fields(s)
	if len(fields) > 0 {
		return fields[0]
	}
	return "0"
}

// ResolveCodec extracts the appropriate FFmpeg encoder string.
func ResolveCodec(c string) string {
	switch {
	case strings.Contains(c, "libx265"):
		return "libx265"
	case strings.Contains(c, "prores"):
		return "prores_ks"
	default:
		return "libx264"
	}
}

// ResolvePreset maps human-readable preset names to FFmpeg presets.
func ResolvePreset(p string) string {
	switch {
	case strings.Contains(p, "veryfast"):
		return "veryfast"
	case strings.Contains(p, "medium"):
		return "medium"
	default:
		return "fast"
	}
}

// ResolveThreads calculates the optimal -j load:proc:save argument for Real-ESRGAN.
func ResolveThreads(mode string) string {
	cpus := runtime.NumCPU()
	switch {
	case strings.Contains(mode, "Standard"):
		return "1:2:1"
	case strings.Contains(mode, "Seimbang") || strings.Contains(mode, "Balanced"):
		return "2:2:2"
	default: // Turbo Multi-Thread / Maksimal
		if cpus >= 16 {
			return "4:4:4"
		} else if cpus >= 8 {
			return "2:4:2"
		}
		return "2:2:2"
	}
}

// FormatDuration formats seconds as MM:SS.S
func FormatDuration(sec float64) string {
	if sec < 0 {
		sec = 0
	}
	totalSec := int(sec)
	m := totalSec / 60
	s := totalSec % 60
	tenths := int(math.Mod(sec, 1.0) * 10)
	return fmt.Sprintf("%02d:%02d.%d", m, s, tenths)
}
