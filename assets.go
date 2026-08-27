package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	var videoInfo struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}
	err = json.Unmarshal(stdout.Bytes(), &videoInfo)
	if err != nil {
		return "", err
	}

	if len(videoInfo.Streams) == 0 {
		return "", nil
	}

	width := videoInfo.Streams[0].Width
	height := videoInfo.Streams[0].Height
	if height == 0 {
		return "", nil
	}

	if width*9 == height*16 {
		return "16:9", nil
	} else if width*16 == height*9 {
		return "9:16", nil
	} else {
		return "other", nil
	}
}
