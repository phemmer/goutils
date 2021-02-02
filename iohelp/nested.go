package iohelp

import (
	"io"
)

type NestedReadCloserT struct {
	io.ReadCloser
	Inner io.ReadCloser
}

// If outer is an io.ReadCloser, close will be called on it first. If outer close errors, inner will not be closed.
func (nrc NestedReadCloserT) Close() error {
	err := nrc.ReadCloser.Close()
	if err == nil {
		err = nrc.Inner.Close()
	}
	return err
}

// Creates a NestedReadCloserT.
//
// If outer is an io.ReadCloser, close will be called on it first. If outer close errors, inner will not be closed.
func NestedReadCloser(outer io.Reader, inner io.ReadCloser) NestedReadCloserT {
	return NestedReadCloserT{
		ReadCloser: ConvReadCloser(outer),
		Inner:      inner,
	}
}

type NestedWriteCloserT struct {
	io.WriteCloser
	Inner io.WriteCloser
}

// Creates a NestedWriteCloserT.
//
// If outer is an io.WriteCloser, close will be called on it first. If outer close errors, inner will not be closed.
func NestedWriteCloser(outer io.Writer, inner io.WriteCloser) NestedWriteCloserT {
	return NestedWriteCloserT{
		WriteCloser: ConvWriteCloser(outer),
		Inner:       inner,
	}
}

// If outer is an io.WriteCloser, close will be called on it first. If outer close errors, inner will not be closed.
func (nwc NestedWriteCloserT) Close() error {
	err := nwc.WriteCloser.Close()
	if err != nil {
		return err
	}
	return nwc.Inner.Close()
}

type nopReadCloser struct {
	io.Reader
}

func (nwc nopReadCloser) Close() error { return nil }

// Ensures r is a io.ReadCloser.
// If r is already a io.ReadCloser, returns r.
// If r is not a io.ReadCloser, returns r wrapped in a type that no-ops Close()
func ConvReadCloser(r io.Reader) io.ReadCloser {
	if r, ok := r.(io.ReadCloser); ok {
		return r
	}
	return nopReadCloser{r}
}

type nopWriteCloser struct {
	io.Writer
}

func (nwc nopWriteCloser) Close() error { return nil }

// Ensures w is a io.WriteCloser.
// If w is already a io.WriteCloser, returns w.
// If w is not a io.WriteCloser, returns w wrapped in a type that no-ops Close()
func ConvWriteCloser(w io.Writer) io.WriteCloser {
	if w, ok := w.(io.WriteCloser); ok {
		return w
	}
	return nopWriteCloser{w}
}
