// Package videos handles different operations on videos
package videos

import (
	"bytes"
	"encoding/json"
	"math"
	"os/exec"
)

const Tolerance = 0.01

type AspectRatio string

const (
	Landscape        AspectRatio = "16:9"
	Portrait         AspectRatio = "9:16"
	AspectRatioOther AspectRatio = "other"
)

// GetVideoAspectRatio returns the aspect ratio of a video in the provided filePath
// the returned aspect ratio is in a string format along wiht an error
func GetVideoAspectRatio(filePath string) (AspectRatio, error) {
	// ffprobe command to show the streams date of the video
	// which is going to be used to get the width and height of the video
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)

	stdOutBuff := bytes.Buffer{}

	cmd.Stdout = &stdOutBuff

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	var data struct {
		Streams []struct {
			Height int `json:"height"`
			Width  int `json:"width"`
		} `json:"streams"`
	}

	err = json.Unmarshal(stdOutBuff.Bytes(), &data)
	if err != nil {
		return "", err
	}

	ratio := float64(data.Streams[0].Width) / float64(data.Streams[0].Height)

	targets := map[AspectRatio]float64{
		Landscape: 16.0 / 9.0,
		Portrait:  9.0 / 16.0,
	}

	for label, target := range targets {
		if math.Abs(ratio-target) <= Tolerance {
			return label, nil
		}
	}

	return AspectRatioOther, nil
}
