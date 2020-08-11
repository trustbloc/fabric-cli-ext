/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	collsCfgJSON               = `[{"name":"dcas","type":"COL_DCAS","policy":"OR('Org1MSP.member','Org2MSP.member')","requiredPeerCount":1,"maxPeerCount":2,"timeToLive":"10m"},{"name":"meta_data","type":"COL_OFFLEDGER","policy":"OR('IMPLICIT-ORG.member')","requiredPeerCount":0,"maxPeerCount":0,"timeToLive":""}]`
	collsCfgJSON_invalidPolicy = `[{"name":"meta_data","type":"COL_OFFLEDGER","policy":"OR('xxx')","requiredPeerCount":0,"maxPeerCount":1,"timeToLive":""}]`
)

func TestUnmarshalCollectionsConfig(t *testing.T) {
	t.Run("No config -> success", func(t *testing.T) {
		cfg, err := UnmarshalCollectionsConfig("")
		require.NoError(t, err)
		require.Empty(t, cfg)
	})

	t.Run("Valid config -> success", func(t *testing.T) {
		cfg, err := UnmarshalCollectionsConfig(collsCfgJSON)
		require.NoError(t, err)
		require.Len(t, cfg, 2)

		collCfg1 := cfg[0].GetStaticCollectionConfig()
		require.NotNil(t, collCfg1)

		collCfg2 := cfg[1].GetStaticCollectionConfig()
		require.NotNil(t, collCfg2)
	})

	t.Run("Unmarshal error -> error", func(t *testing.T) {
		cfg, err := UnmarshalCollectionsConfig("[")
		require.EqualError(t, err, "invalid collections config: unexpected end of JSON input")
		require.Empty(t, cfg)
	})

	t.Run("Invalid policy -> error", func(t *testing.T) {
		cfg, err := UnmarshalCollectionsConfig(collsCfgJSON_invalidPolicy)
		require.Error(t, err)
		require.Contains(t, err.Error(), `unrecognized token 'xxx' in policy string`)
		require.Empty(t, cfg)
	})
}
