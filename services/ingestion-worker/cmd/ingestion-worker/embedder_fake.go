package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

type Embedder interface {
	Embed(text string) ([]float32, error)
	Dim() int
}

type FakeEmbedder struct {
	dim int
}

func NewFakeEmbedder(dim int) *FakeEmbedder {
	if dim <= 0 {
		dim = 8
	}
	return &FakeEmbedder{dim: dim}
}

func (e *FakeEmbedder) Dim() int { return e.dim }

func (e *FakeEmbedder) Embed(text string) ([]float32, error) {
	sum := sha256.Sum256([]byte(text))

	// Deterministic pseudo-random-ish values in [-1, 1].
	out := make([]float32, e.dim)
	for i := 0; i < e.dim; i++ {
		off := (i * 4) % (len(sum) - 3)
		u := binary.LittleEndian.Uint32(sum[off : off+4])
		out[i] = (float32(u)/float32(^uint32(0)))*2 - 1
	}

	if len(out) != e.dim {
		return nil, fmt.Errorf("fake embedder produced wrong dim")
	}
	return out, nil
}
