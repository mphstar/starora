package engine

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// FormatBytes formats a byte count into a human-readable size string.
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// ResolveVideoCRF parses the target CRF value from the selected preset.
func ResolveVideoCRF(crfString string, fallbackIndex int) string {
	switch {
	case strings.Contains(crfString, "28"):
		return "28"
	case strings.Contains(crfString, "26"):
		return "26"
	case strings.Contains(crfString, "23"):
		return "23"
	case strings.Contains(crfString, "22"):
		return "22"
	case strings.Contains(crfString, "18"):
		return "18"
	case strings.Contains(crfString, "17"):
		return "17"
	default:
		switch fallbackIndex {
		case 0:
			return "17"
		case 2:
			return "26"
		default:
			return "21"
		}
	}
}

// CompressImageWithFFmpeg compresses an image file using FFmpeg while preserving desired format.
// Default output is PNG with optimal compression level 9 and mixed prediction.
func CompressImageWithFFmpeg(ctx context.Context, ffmpegPath, inPath, outPath, mode string, quality int) error {
	var args []string
	ext := strings.ToLower(filepath.Ext(inPath))

	switch {
	case strings.Contains(mode, "WebP") || ext == ".webp":
		q := quality
		if q <= 0 || q > 100 {
			q = 82
		}
		args = []string{"-y", "-threads", "0", "-i", inPath, "-c:v", "libwebp", "-quality", fmt.Sprintf("%d", q), outPath}
	case strings.Contains(mode, "JPG") || ext == ".jpg" || ext == ".jpeg":
		args = []string{"-y", "-threads", "0", "-i", inPath, "-q:v", "2", outPath}
	default: // PNG (Default format)
		args = []string{"-y", "-threads", "0", "-i", inPath, "-c:v", "png", "-compression_level", "9", "-pred", "mixed", outPath}
	}

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	SetHideWindow(cmd)
	return cmd.Run()
}
