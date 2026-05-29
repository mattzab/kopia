package compression

import (
	"io"

	"github.com/ulikunitz/xz"
	"github.com/pkg/errors"

	"github.com/kopia/kopia/internal/iocopy"
)

func init() {
	RegisterCompressor("xz", newXZCompressor(headerXZDefault, xzDefaultConfig()))
	RegisterCompressor("xz-best-compression", newXZCompressor(headerXZBestCompression, xzBestConfig()))
}

func xzDefaultConfig() xz.WriterConfig {
	return xz.WriterConfig{
		DictCap:  8 * 1024 * 1024,
		CheckSum: xz.CRC64,
	}
}

func xzBestConfig() xz.WriterConfig {
	return xz.WriterConfig{
		DictCap:  1610612736,
		CheckSum: xz.CRC64,
	}
}

func newXZCompressor(id HeaderID, cfg xz.WriterConfig) Compressor {
	mustSucceed(cfg.Verify())

	return &xzCompressor{id, compressionHeader(id), cfg}
}

type xzCompressor struct {
	id     HeaderID
	header []byte
	cfg    xz.WriterConfig
}

func (c *xzCompressor) HeaderID() HeaderID {
	return c.id
}

func (c *xzCompressor) Compress(output io.Writer, input io.Reader) error {
	if _, err := output.Write(c.header); err != nil {
		return errors.Wrap(err, "unable to write header")
	}

	w, err := c.cfg.NewWriter(output)
	if err != nil {
		return errors.Wrap(err, "error creating xz writer")
	}

	if err := iocopy.JustCopy(w, input); err != nil {
		return errors.Wrap(err, "compression error")
	}

	if err := w.Close(); err != nil {
		return errors.Wrap(err, "compression close error")
	}

	return nil
}

func (c *xzCompressor) Decompress(output io.Writer, input io.Reader, withHeader bool) error {
	if withHeader {
		if err := verifyCompressionHeader(input, c.header); err != nil {
			return err
		}
	}

	r, err := xz.NewReader(input)
	if err != nil {
		return errors.Wrap(err, "error creating xz reader")
	}

	if err := iocopy.JustCopy(output, r); err != nil {
		return errors.Wrap(err, "decompression error")
	}

	return nil
}