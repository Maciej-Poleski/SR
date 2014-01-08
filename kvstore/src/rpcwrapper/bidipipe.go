package rpcwrapper

import (
	"io"
)

type bidiPipe struct {
	*io.PipeReader
	*io.PipeWriter
}

func (pipe bidiPipe) Close() error {
	err := pipe.PipeReader.Close()
	if err1 := pipe.PipeWriter.Close(); err == nil {
		err = err1
	}
	return err
}

func newBidiPipe() (bidiPipe, bidiPipe) {
	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()
	return bidiPipe{r1, w2}, bidiPipe{r2, w1}
}
