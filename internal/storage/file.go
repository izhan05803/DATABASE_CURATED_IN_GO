package storage

import (
	"encoding/binary"
	"fmt"
	"os"
)

const (
	MagicNumber = "GODB"
	HeaderSize  = 100
	Version     = 1
)

// FileHeader represents the database file header
type FileHeader struct {
	Magic        [4]byte
	Version      uint32
	PageSize     uint32
	TotalPages   uint32
	FreeListHead uint32
	RootPage     uint32
}

// FileManager handles file I/O operations
type FileManager struct {
	file   *os.File
	header FileHeader
}

// NewFileManager creates or opens a database file
func NewFileManager(path string) (*FileManager, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	fm := &FileManager{file: file}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("stat file: %w", err)
	}

	if info.Size() == 0 {
		if err := fm.initializeFile(); err != nil {
			file.Close()
			return nil, fmt.Errorf("initialize file: %w", err)
		}
	} else {
		if err := fm.readHeader(); err != nil {
			file.Close()
			return nil, fmt.Errorf("read header: %w", err)
		}
	}

	return fm, nil
}

// initializeFile creates the initial file structure
func (fm *FileManager) initializeFile() error {
	copy(fm.header.Magic[:], MagicNumber)
	fm.header.Version = Version
	fm.header.PageSize = PageSize
	fm.header.TotalPages = 0
	fm.header.FreeListHead = 0
	fm.header.RootPage = 0

	return fm.writeHeader()
}

// writeHeader writes the header to disk
func (fm *FileManager) writeHeader() error {
	if _, err := fm.file.Seek(0, 0); err != nil {
		return err
	}
	return binary.Write(fm.file, binary.LittleEndian, &fm.header)
}

// readHeader reads the header from disk
func (fm *FileManager) readHeader() error {
	if _, err := fm.file.Seek(0, 0); err != nil {
		return err
	}
	return binary.Read(fm.file, binary.LittleEndian, &fm.header)
}

// Close closes the file
func (fm *FileManager) Close() error {
	if fm.file != nil {
		return fm.file.Close()
	}
	return nil
}
