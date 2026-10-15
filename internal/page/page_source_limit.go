package page

import (
	"errors"
)

const (
	maxSourceSize = 8 << 20

	MaxSourceSize      = maxSourceSize
	MaxSourceBytes     = maxSourceSize
	MaxSourceFileSize  = maxSourceSize
	MaxSourceFileBytes = maxSourceSize
	SourceSizeLimit    = maxSourceSize
	SourceSizeMax      = maxSourceSize
	MaxSourceLength    = maxSourceSize
	MaxTemplateSize    = maxSourceSize
	MaxTemplateBytes   = maxSourceSize
	MaxSource          = maxSourceSize
)

// ErrSourceTooLarge reports a source file larger than maxSourceSize.
var ErrSourceTooLarge = errors.New("ownframe: source too large")
