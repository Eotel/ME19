package fileio

import (
	"os"
	"sync"
)

// Writer handles writing QR code data to files
type Writer struct {
	filePath string
	mutex    sync.Mutex
}

// New creates a new file writer
func New(filePath string) *Writer {
	return &Writer{
		filePath: filePath,
	}
}

// WriteData writes the given QR code data to the file
func (w *Writer) WriteData(data string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// File writing will be implemented here
	// For now, this is a placeholder
	file, err := os.OpenFile(w.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(data)
	return err
}
