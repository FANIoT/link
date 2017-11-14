/*
 * +===============================================
 * | Author:        Parham Alvani <parham.alvani@gmail.com>
 * |
 * | Creation Date: 13-11-2017
 * |
 * | File Name:     decoder.go
 * +===============================================
 */

package decoder

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Decoder is a way to communicate with user provided decoder
type Decoder struct {
	addr string
}

// New creates new decoder based on given remove address
func New(addr string) *Decoder {
	return &Decoder{
		addr: addr,
	}
}

// Decode decodes given data with user provided decoder
func (d *Decoder) Decode(payload []byte, id string) (string, error) {
	r, err := http.Post(fmt.Sprintf("%s/api/decode/%s", d.addr, id), "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	switch r.StatusCode {
	case 200:
		return string(b), nil
	case 404:
		return "", fmt.Errorf("%s", b)
	default:
		return "", fmt.Errorf("unkown status code \"%s\"", b)
	}
}