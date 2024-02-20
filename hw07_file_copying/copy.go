package main

import (
	"errors"
	"fmt"
	"os"

	pb "github.com/cheggaaa/pb/v3"
)

var (
	ErrSourceFileNotFound      = errors.New("source file not found")
	ErrReadingFromSource       = errors.New("error reading from source file")
	ErrCreatingDestinationFile = errors.New("could not create destination file")
	ErrUnsupportedFile         = errors.New("unsupported file")
	ErrOffsetExceedsFileSize   = errors.New("offset exceeds file size")
	ErrNegativeOffset          = errors.New("offset cannot be negative")
	ErrNegativeLimit           = errors.New("limit cannot be negative")
	ErrIncompleteCopy          = errors.New("number of copied bytes is less then expected")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	fromFile, err := os.Open(fromPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = ErrSourceFileNotFound
			fmt.Println(err)
		}
		return err
	}
	defer fromFile.Close()

	toFile, err := os.Create(toPath)
	if err != nil {
		return ErrCreatingDestinationFile
	}
	defer toFile.Close()

	fileInfo, err := os.Stat(fromPath)
	if err != nil {
		err = ErrUnsupportedFile
		fmt.Println(err)
		return err
	}
	fileSize := fileInfo.Size()

	switch {
	case offset > fileSize:
		return ErrOffsetExceedsFileSize
	case offset < 0:
		return ErrNegativeOffset
	case limit < 0:
		return ErrNegativeLimit
	case (limit == 0) || (offset == 0 && limit >= fileSize):
		offset = 0
		limit = fileSize
		fallthrough
	default:
		var bufSizeMax int64
		bufSizeMax = 1024
		if limit >= (fileSize - offset) {
			limit = fileSize - offset
		}

		if limit < 1024 {
			bufSizeMax = limit
		}

		var bytesCopied int64
		var bufSize int64
		bufSize = bufSizeMax
		buf := make([]byte, bufSizeMax)
		bar := pb.Full.Start64(limit)
		for (limit - bytesCopied) > 0 {
			if (limit - bytesCopied) < bufSizeMax {
				bufSize = limit - bytesCopied
			}

			read, err := fromFile.ReadAt(buf[:bufSize], offset)
			if err != nil {
				return ErrReadingFromSource
			}

			written, err := toFile.WriteAt(buf[:bufSize], bytesCopied)
			if err != nil || written < read {
				return ErrIncompleteCopy
			}
			bytesCopied += int64(read)
			offset += int64(read)
			bar.SetCurrent(bytesCopied)
		}

		bar.Finish()
	}

	return err
}
