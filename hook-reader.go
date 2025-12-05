package rustfs

import (
	"fmt"
	"io"
)

// hookReader hooks additional reader in the source stream. It is
// useful for making progress bars. Second reader is appropriately
// notified about the exact number of bytes read from the primary
// source on each Read operation.
type hookReader struct {
	source io.Reader
	hook   io.Reader
}

// Seek implements io.Seeker. Seeks source first, and if necessary
// seeks hook if Seek method is appropriately found.
func (hr *hookReader) Seek(offset int64, whence int) (n int64, err error) {
	// Verify for source has embedded Seeker, use it.
	sourceSeeker, ok := hr.source.(io.Seeker)
	if ok {
		n, err = sourceSeeker.Seek(offset, whence)
		if err != nil {
			return 0, err
		}
	}

	if hr.hook != nil {
		// Verify if hook has embedded Seeker, use it.
		hookSeeker, ok := hr.hook.(io.Seeker)
		if ok {
			var m int64
			m, err = hookSeeker.Seek(offset, whence)
			if err != nil {
				return 0, err
			}
			if n != m {
				return 0, fmt.Errorf("hook seeker seeked %d bytes, expected source %d bytes", m, n)
			}
		}
	}

	return n, nil
}

// Read implements io.Reader. Always reads from the source, the return
// value 'n' number of bytes are reported through the hook. Returns
// error for all non io.EOF conditions.
func (hr *hookReader) Read(b []byte) (n int, err error) {
	n, err = hr.source.Read(b)
	if err != nil && err != io.EOF {
		return n, err
	}
	if hr.hook != nil {
		// Progress the hook with the total read bytes from the source.
		if _, herr := hr.hook.Read(b[:n]); herr != nil {
			if herr != io.EOF {
				return n, herr
			}
		}
	}
	return n, err
}

// newHook returns a io.ReadSeeker which implements hookReader that
// reports the data read from the source to the hook.
func newHook(source, hook io.Reader) io.Reader {
	if hook == nil {
		return &hookReader{source: source}
	}
	return &hookReader{
		source: source,
		hook:   hook,
	}
}
