/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

// Format specifies the format of the configuration
type Format string

// Config contains zero or more application configurations and zero or more peer-specific application configurations
type Config struct {
	// MspID is the ID of the MSP
	MspID string
	// Peers contains configuration for zero or more peers
	Peers []*Peer `json:",omitempty"`
	// Apps contains configuration for zero or more application
	Apps []*App `json:",omitempty"`
}

// Peer contains a collection of application configurations for a given peer
type Peer struct {
	// PeerID is the unique ID of the peer
	PeerID string
	// Apps contains configuration for one or more application
	Apps []*App
}

// App contains the configuration for an application and/or multiple sub-components.
type App struct {
	// Name is the name of the application
	AppName string
	// Version is the version of the config
	Version string
	// Format describes the format of the data
	Format Format
	// Config contains the actual configuration
	Config string
	// Components zero or more component configs
	Components []*Component `json:",omitempty"`
}

// Component contains the configuration for an application component.
type Component struct {
	// Name is the name of the component
	Name string
	// Version is the version of the config
	Version string
	// Format describes the format of the data
	Format Format
	// Config contains the actual configuration
	Config string
}
