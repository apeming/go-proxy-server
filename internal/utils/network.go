package utils

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"go-proxy-server/internal/constants"
	"go-proxy-server/internal/logger"
)

// BufferPool is a pool of byte buffers to reduce GC pressure
// Uses large buffer size for bulk data transfers
var BufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, constants.BufferSizeLarge)
	},
}

// CopyWithIdleTimeout copies data from src to dst with idle timeout
// It resets the deadline after each successful read/write operation
// Uses buffer pool to reduce GC pressure
func CopyWithIdleTimeout(ctx context.Context, dst, src net.Conn, readTimeout, writeTimeout time.Duration) error {
	// Get buffer from pool
	buf := BufferPool.Get().([]byte)
	defer BufferPool.Put(buf) // Return buffer to pool when done

	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Set read deadline (idle timeout)
		src.SetReadDeadline(time.Now().Add(readTimeout))

		n, err := src.Read(buf)
		if n > 0 {
			// Set write deadline (idle timeout)
			dst.SetWriteDeadline(time.Now().Add(writeTimeout))

			_, writeErr := dst.Write(buf[:n])
			if writeErr != nil {
				return writeErr
			}
		}

		if err != nil {
			if err == io.EOF {
				return nil
			}
			// Check if it's a timeout error
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				logger.Warn("Idle timeout reached during data transfer")
			}
			return err
		}
	}
}
