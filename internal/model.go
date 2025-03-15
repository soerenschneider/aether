package internal

import "errors"

var ErrTemplate = errors.New("template error")

type Data struct {
	Summary                    []string
	RenderedDefaultTemplate    []byte
	RenderedSimplifiedTemplate []byte
}
