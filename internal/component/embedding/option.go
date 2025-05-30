package embedding

import "time"

type EmbeddingOption func(*EmbeddingOptions)

type EmbeddingOptions struct {
	Timeout *time.Duration
}

func WithTimeout(timeout time.Duration) EmbeddingOption {
	return func(o *EmbeddingOptions) {
		o.Timeout = &timeout
	}
}
