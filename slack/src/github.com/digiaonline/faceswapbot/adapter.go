package faceswapbot

import "os"

type FaceReplacer interface {
	SwapFaces(orig *os.File) (*os.File, error)
	PhotoBomb(orig *os.File) (*os.File, error)
	Success(orig *os.File) (*os.File, error)
}
