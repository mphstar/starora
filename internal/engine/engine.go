package engine

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"starora/embedded"
)

func SetHideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
}

type EnginePaths struct {
	RealESRGAN string
	ModelsDir  string
	FFmpeg     string
	RootDir    string
}

func GetEngineRootDir() string {
	appData, err := os.UserConfigDir()
	if err != nil {
		appData = os.TempDir()
	}
	return filepath.Join(appData, "Starora", "engine")
}

func EnsureEmbeddedEngines() (EnginePaths, error) {
	engineDir := GetEngineRootDir()
	if err := os.MkdirAll(engineDir, 0755); err != nil {
		return EnginePaths{}, err
	}

	realesrganExe := filepath.Join(engineDir, "realesrgan-ncnn-vulkan.exe")
	ffmpegExe := filepath.Join(engineDir, "ffmpeg.exe")
	modelsDir := filepath.Join(engineDir, "models")

	modelsExist := false
	if entries, err := os.ReadDir(modelsDir); err == nil && len(entries) > 0 {
		modelsExist = true
	}

	if _, err := os.Stat(realesrganExe); os.IsNotExist(err) || !modelsExist {
		if err := unzipBytes(embedded.RealESRGANZip, engineDir); err != nil {
			return EnginePaths{}, fmt.Errorf("failed to extract Real-ESRGAN engine: %w", err)
		}
	}

	if _, err := os.Stat(ffmpegExe); os.IsNotExist(err) {
		if err := unzipBytes(embedded.FFmpegZip, engineDir); err != nil {
			return EnginePaths{}, fmt.Errorf("failed to extract FFmpeg engine: %w", err)
		}
	}

	return EnginePaths{
		RealESRGAN: realesrganExe,
		ModelsDir:  modelsDir,
		FFmpeg:     ffmpegExe,
		RootDir:    engineDir,
	}, nil
}

func unzipBytes(zipData []byte, destDir string) error {
	r, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return err
	}

	for _, f := range r.File {
		outPath := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(outPath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
