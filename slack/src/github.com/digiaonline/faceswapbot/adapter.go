package faceswapbot

import "os"

type FaceReplacer interface {
	// Swap faces or photobomb if none found
	// use bw to indicate if the output should be grayscale/b&w
	SwapFaces(orig *os.File, bw bool) (*os.File, error)
	// Explicitly photobomb the image
	// use bw to indicate if the output should be grayscale/b&w
	PhotoBomb(orig *os.File, bw bool) (*os.File, error)
	// Use "Success" photobomb image
	// use bw to indicate if the output should be grayscale/b&w
	Success(orig *os.File, bw bool) (*os.File, error)
}
