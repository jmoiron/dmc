package main

import (
	"io"
	"sync"
)

// A SyncLineWriter is a LineWriter that is safe to use across goroutines.
type SyncLineWriter struct {
	out io.Writer
	buf chan string
	wg  sync.WaitGroup
}

// NewSyncLineWriter returns a SyncLineWriter that writes its lines out to the
// provided writer.  It uses one goroutine, which is cleaned up when calling
// Close.
func NewSyncLineWriter(w io.Writer) *SyncLineWriter {
	s := &SyncLineWriter{out: w, buf: make(chan string, 512)}
	s.run()
	return s
}

func (s *SyncLineWriter) run() {
	s.wg.Add(1)
	go func() {
		for l := range s.buf {
			s.out.Write([]byte(l))
		}
		s.wg.Done()
	}()
}

// Close flushes the rest of the output and closes the writer.
// It is not legal to write to this LineBufferedWriter after closing.
func (s *SyncLineWriter) Close() error {
	close(s.buf)
	s.wg.Wait()
	return nil
}

// WriteLine writes the str to this writer.
func (s *SyncLineWriter) WriteLine(str string) error {
	s.buf <- str
	return nil
}
