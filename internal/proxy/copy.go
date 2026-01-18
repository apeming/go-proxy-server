package proxy

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"go-proxy-server/internal/constants"
	"go-proxy-server/internal/logger"
)

// bufferPool is a pool of byte buffers to reduce GC pressure
// Uses large buffer size for bulk data transfers
var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, constants.BufferSizeLarge)
	},
}

// copyWithIdleTimeout copies data from src to dst with idle timeout
// It resets the deadline after each successful read/write operation
// Uses buffer pool to reduce GC pressure
func copyWithIdleTimeout(ctx context.Context, dst, src net.Conn, readTimeout, writeTimeout time.Duration) error {
	// Get buffer from pool
	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf) // Return buffer to pool when done

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

			// Ensure all data is written (handle partial writes)
			written := 0
			for written < n {
				nw, writeErr := dst.Write(buf[written:n])
				if writeErr != nil {
					return writeErr
				}
				written += nw
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
