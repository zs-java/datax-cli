package libdatax

import "io"

type MultipleWriter struct {
	writers []io.Writer
}

func NewMultipleWriter(writers ...io.Writer) *MultipleWriter {
	return &MultipleWriter{writers: writers}
}

func (c *MultipleWriter) Write(p []byte) (n int, err error) {
	for _, w := range c.writers {
		n, err = w.Write(p)
		if err != nil {
			break
		}
	}
	return n, err
}
