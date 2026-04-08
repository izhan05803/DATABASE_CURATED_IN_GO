package storage

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/izhan05803/gofromscratchdb/pkg/types"
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

	buf := make([]byte, HeaderSize)
	copy(buf[0:4], fm.header.Magic[:])
	binary.LittleEndian.PutUint32(buf[4:8], fm.header.Version)
	binary.LittleEndian.PutUint32(buf[8:12], fm.header.PageSize)
	binary.LittleEndian.PutUint32(buf[12:16], fm.header.TotalPages)
	binary.LittleEndian.PutUint32(buf[16:20], fm.header.FreeListHead)
	binary.LittleEndian.PutUint32(buf[20:24], fm.header.RootPage)

	_, err := fm.file.Write(buf)
	return err
}

// readHeader reads the header from disk
func (fm *FileManager) readHeader() error {
	if _, err := fm.file.Seek(0, 0); err != nil {
		return err
	}

	buf := make([]byte, HeaderSize)
	if _, err := fm.file.Read(buf); err != nil {
		return err
	}

	copy(fm.header.Magic[:], buf[0:4])
	if string(fm.header.Magic[:]) != MagicNumber {
		return types.ErrInvalidFormat
	}

	fm.header.Version = binary.LittleEndian.Uint32(buf[4:8])
	fm.header.PageSize = binary.LittleEndian.Uint32(buf[8:12])
	fm.header.TotalPages = binary.LittleEndian.Uint32(buf[12:16])
	fm.header.FreeListHead = binary.LittleEndian.Uint32(buf[16:20])
	fm.header.RootPage = binary.LittleEndian.Uint32(buf[20:24])

	return nil
}

// WritePayload writes serialized records right after the fixed-size header.
func (fm *FileManager) WritePayload(payload []byte) error {
	if _, err := fm.file.Seek(HeaderSize, io.SeekStart); err != nil {
		return err
	}

	if err := fm.file.Truncate(HeaderSize); err != nil {
		return err
	}

	if len(payload) == 0 {
		return fm.file.Sync()
	}

	if _, err := fm.file.Write(payload); err != nil {
		return err
	}

	return fm.file.Sync()
}

// ReadPayload reads serialized records from the database file.
func (fm *FileManager) ReadPayload() ([]byte, error) {
	if _, err := fm.file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	info, err := fm.file.Stat()
	if err != nil {
		return nil, err
	}

	if info.Size() <= HeaderSize {
		return []byte{}, nil
	}

	payloadSize := info.Size() - HeaderSize
	payload := make([]byte, payloadSize)

	if _, err := fm.file.Seek(HeaderSize, io.SeekStart); err != nil {
		return nil, err
	}

	if _, err := io.ReadFull(fm.file, payload); err != nil {
		return nil, err
	}

	return payload, nil
}

// Close closes the file
func (fm *FileManager) Close() error {
	if fm.file != nil {
		return fm.file.Close()
	}
	return nil
}
