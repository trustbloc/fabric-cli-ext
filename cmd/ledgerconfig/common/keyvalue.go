/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"fmt"
)

// Key is used to uniquely identify a specific application configuration and is used as the
// key when persisting to a state store.
type Key struct {
	// MspID is the ID of the MSP that owns the data
	MspID string
	// PeerID is the (optional) ID of the peer with which the data is associated
	PeerID string `json:",omitempty"`
	// AppName is the name of the application that owns the data
	AppName string
	// AppVersion is the version of the application config
	AppVersion string
	// ComponentName is the (optional) name of the application component
	ComponentName string `json:",omitempty"`
	// ComponentVersion is the (optional) version of the application component config
	ComponentVersion string `json:",omitempty"`
}

// String returns a readable string for the key
func (k *Key) String() string {
	return fmt.Sprintf("(MSP:%s),(Peer:%s),(AppName:%s),(AppVersion:%s),(Comp:%s),(CompVersion:%s)", k.MspID, k.PeerID, k.AppName, k.AppVersion, k.ComponentName, k.ComponentVersion)
}

// Value contains the configuration data and is persisted as a JSON document in the store.
type Value struct {
	// TxID is the ID of the transaction in which the config was stored/updated
	TxID string
	// Format describes the format (type) of the config data
	Format Format
	// Config contains the actual configuration
	Config string
	// Tags contains an optional set of tags that describe the data
	Tags []string
}

// String returns a readable string for the value
func (v *Value) String() string {
	return fmt.Sprintf("(TxID:%s),(Config:%s),(Format:%s),(Tags:%s)", v.TxID, v.Config, v.Format, v.Tags)
}

// KeyValue contains the key and the value for the key
type KeyValue struct {
	*Key
	*Value
}

// String returns a readable string for the key-value
func (kv *KeyValue) String() string {
	return fmt.Sprintf("[%s]=[%s]", kv.Key, kv.Value)
}
