/*
 *
 * In The Name of God
 *
 * +===============================================
 * | Author:        Parham Alvani <parham.alvani@gmail.com>
 * |
 * | Creation Date: 14-09-2018
 * |
 * | File Name:     protocol.go
 * +===============================================
 */

package app

import "github.com/I1820/types"

// Protocol is a way for getting data from lower levels -- MQTT, etc. --
// then parses and creates a Data instance.
// Parse means extract data, link quality (if available), etc. from the raw message.
type Protocol interface {
	TxTopic() string
	RxTopic() string

	Name() string

	Marshal([]byte) (types.Data, error)
}
