// Package object object/checksum.go
package object

import "github.com/Scorpio69t/rustfs-go/internal/core"

func applyChecksumHeaders(meta *core.RequestMetadata, options PutOptions) {
	if options.ChecksumMode != "" {
		meta.CustomHeader.Set("x-amz-checksum-mode", options.ChecksumMode)
	}
	if options.ChecksumAlgorithm != "" {
		meta.CustomHeader.Set("x-amz-checksum-algorithm", options.ChecksumAlgorithm)
	}
}
