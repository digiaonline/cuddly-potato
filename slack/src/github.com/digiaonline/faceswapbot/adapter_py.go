package faceswapbot

import (
	"os"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"bytes"
	"log"
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

	// The python face_swapper does not support outputting gifs... yet
	fileExt := filepath.Ext(f.Name())
	if fileExt == ".gif" {
		fileExt = ".png"
	}

	return tmpFile.Name() + fileExt, nil
}

// Run the given command and get the stdout and stderror as strings
func runCommand(cmd *exec.Cmd) (stdout, stderr string, err error) {
	var outbuf, errbuf bytes.Buffer

	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err = cmd.Run()
	stdout = outbuf.String()
	stderr = errbuf.String()

	return
}

// Swaps any faces found in the original image
// Photobomb if no faces found (implementation specifics)
func (a PySwapper) SwapFaces(orig *os.File, bw bool) (*os.File, error) {
	outName, err := getTempFileName(orig)
	if err != nil {
		return nil, err
	}

	args := []string{
		a.Executable,
		orig.Name(),
		"-f",
		a.FacesPath,
		"-b",
		a.BodiesPath,
		"-o",
		outName,
	}

	if true == bw {
		args = append(args, "-bw")
	}

	cmd := exec.Command(
		"python",
		args...
	)

	_, stdErr, err := runCommand(cmd)
	if err != nil {
		log.Printf(stdErr)
		return nil, err
	}

	outFile, err := os.Open(outName)
	if err != nil {
		return nil, err
	}

	return outFile, nil
}

// Photobomb image regardless of found faces
func (a PySwapper) PhotoBomb(orig *os.File, bw bool) (*os.File, error) {
	outName, err := getTempFileName(orig)
	if err != nil {
		return nil, err
	}

	args := []string{
		a.Executable,
		orig.Name(),
		"-f",
		a.FacesPath,
		"-b",
		a.BodiesPath,
		"-p", // photobomb
		"-o",
		outName,
	}

	if true == bw {
		args = append(args, "-bw")
	}

	cmd := exec.Command(
		"python",
		args...
	)

	_, stdErr, err := runCommand(cmd)
	if err != nil {
		log.Printf(stdErr)
		return nil, err
	}

	outFile, err := os.Open(outName)
	if err != nil {
		return nil, err
	}

	return outFile, nil
}

// Photobomb with the success image
func (a PySwapper) Success(orig *os.File, bw bool) (*os.File, error) {
	outName, err := getTempFileName(orig)
	if err != nil {
		return nil, err
	}

	args := []string{
		a.Executable,
		orig.Name(),
		"-f",
		a.FacesPath,
		"-b",
		a.SuccessPath, // the success path should be a path to a single image
		"-p",          // For the "success" to succeed, it needs the photobomb flag
		"-o",
		outName,
	}

	if true == bw {
		args = append(args, "-bw")
	}

	cmd := exec.Command(
		"python",
		args...
	)
	_, stdErr, err := runCommand(cmd)
	if err != nil {
		log.Printf(stdErr)
		return nil, err
	}

	outFile, err := os.Open(outName)
	if err != nil {
		return nil, err
	}

	return outFile, nil
}
