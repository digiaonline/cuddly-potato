package faceswapbot

import (
	"os"
	"io/ioutil"
	"os/exec"
	"path/filepath"
)

type PySwapper struct {
	Executable  string
	FacesPath   string
	BodiesPath  string
	SuccessPath string
}

// A stupid method to get a like suffixed temporary files name
func getTempFileName(f *os.File) (string, error) {
	tmpFile, err := ioutil.TempFile("", "slack_image")
	if err != nil {
		return "", err
	}

	defer os.Remove(tmpFile.Name())

	return tmpFile.Name() + filepath.Ext(f.Name()), nil
}

// Swaps any faces found in the original image
// Photobomb if no faces found (implementation specifics)
func (a PySwapper) SwapFaces(orig *os.File) (*os.File, error) {
	outName, err := getTempFileName(orig)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(
		"python",
		a.Executable,
		orig.Name(),
		"-f",
		a.FacesPath,
		"-b",
		a.BodiesPath,
		"-o",
		outName,
	)

	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	outFile, err := os.Open(outName)
	if err != nil {
		return nil, err
	}

	return outFile, nil
}

// Photobomb image regardless of found faces
func (a PySwapper) PhotoBomb(orig *os.File) (*os.File, error) {
	outName, err := getTempFileName(orig)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(
		"python",
		a.Executable,
		orig.Name(),
		"-f",
		a.FacesPath,
		"-b",
		a.BodiesPath,
		"-p", // photobomb
		"-o",
		outName,
	)

	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	outFile, err := os.Open(outName)
	if err != nil {
		return nil, err
	}

	return outFile, nil
}

// Photobomb with the success image
func (a PySwapper) Success(orig *os.File) (*os.File, error) {
	outName, err := getTempFileName(orig)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(
		"python",
		a.Executable,
		orig.Name(),
		"-f",
		a.FacesPath,
		"-b",
		a.SuccessPath, // the success path should be a path to a single image
		"-p",          // For the "success" to succeed, it needs the photobomb flag
		"-o",
		outName,
	)

	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	outFile, err := os.Open(outName)
	if err != nil {
		return nil, err
	}

	return outFile, nil
}
