/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

const (
	// ConfigSCC is the name of the ledger configuration system chaincode
	ConfigSCC = "configscc"
)

// Criteria is used for configuration searches
type Criteria struct {
	// MspID is the ID of the MSP that owns the data
	MspID string `json:",omitempty"`

	// PeerID is the ID of the peer with which the data is associated
	PeerID string `json:",omitempty"`

	// AppName is the name of the application that owns the data
	AppName string `json:",omitempty"`

	// AppVersion is the version of the application config
	AppVersion string `json:",omitempty"`

	// ComponentName is the name of the application component
	ComponentName string `json:",omitempty"`

	// ComponentVersion is the version of the application component config
	ComponentVersion string `json:",omitempty"`
}
