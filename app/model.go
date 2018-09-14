/*
 *
 * In The Name of God
 *
 * +===============================================
 * | Author:        Parham Alvani <parham.alvani@gmail.com>
 * |
 * | Creation Date: 14-09-2018
 * |
 * | File Name:     model.go
 * +===============================================
 */

package app

// Model is a decoder/encoder interface.
// It specifies a way for creating useful information
// from raw data that are coming from devices.
// examples contain generic (based on user scripts) or aolab
type Model interface {
	Decode([]byte) interface{}
	Encode(interface{}) []byte

	Name() string
}
