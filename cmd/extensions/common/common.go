/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"encoding/json"

	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/policydsl"
	"github.com/pkg/errors"
)

// UnmarshalCollectionsConfig unmarshals the given collections config.
func UnmarshalCollectionsConfig(collsConfig string) ([]*pb.CollectionConfig, error) {
	if collsConfig == "" {
		return nil, nil
	}

	var cconf []collectionConfigJSON
	if err := json.Unmarshal([]byte(collsConfig), &cconf); err != nil {
		return nil, errors.WithMessagef(err, "invalid collections config")
	}

	ccarray := make([]*pb.CollectionConfig, 0, len(cconf))
	for _, cconfitem := range cconf {
		p, err := policydsl.FromString(cconfitem.Policy)
		if err != nil {
			return nil, err
		}
		cpc := &pb.CollectionPolicyConfig{
			Payload: &pb.CollectionPolicyConfig_SignaturePolicy{
				SignaturePolicy: p,
			},
		}

		cc := &pb.CollectionConfig{
			Payload: &pb.CollectionConfig_StaticCollectionConfig{
				StaticCollectionConfig: &pb.StaticCollectionConfig{
					Name:              cconfitem.Name,
					Type:              pb.CollectionType(pb.CollectionType_value[cconfitem.Type]),
					MemberOrgsPolicy:  cpc,
					RequiredPeerCount: cconfitem.RequiredCount,
					MaximumPeerCount:  cconfitem.MaxPeerCount,
					BlockToLive:       cconfitem.BlockToLive,
					TimeToLive:        cconfitem.TimeToLive,
				},
			},
		}
		ccarray = append(ccarray, cc)
	}
	return ccarray, nil
}

type collectionConfigJSON struct {
	Name            string `json:"name"`
	Type            string `json:"type"`
	Policy          string `json:"policy"`
	RequiredCount   int32  `json:"requiredPeerCount"`
	MaxPeerCount    int32  `json:"maxPeerCount"`
	BlockToLive     uint64 `json:"blockToLive"`
	TimeToLive      string `json:"timeToLive"`
	MemberOnlyRead  bool   `json:"memberOnlyRead"`
	MemberOnlyWrite bool   `json:"memberOnlyWrite"`
}
