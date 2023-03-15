/*
Copyright 2022-2023 The Nephio Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"fmt"

	"github.com/nephio-seed/nf-deploy-controller/hydration/types"
)

func GetInterfaceConfigSpecMap(configs []*types.InterfaceConfig) map[int]types.InterfaceCfgSpec {
	resp := make(map[int]types.InterfaceCfgSpec)
	for _, config := range configs {
		for _, spec := range config.Spec {
			resp[spec.ID] = spec
		}
	}
	return resp
}

func GetBGPConfig(nfBgpConfig *types.NFBGPConfig) []types.BGPConfig {
	if nfBgpConfig == nil {
		return nil
	}
	nfBgpSpec := nfBgpConfig.Spec
	resp := make([]types.BGPConfig, len(nfBgpSpec.VRouters))
	for i, vRouter := range nfBgpSpec.VRouters {
		resp[i] = types.BGPConfig{
			VRName:             vRouter.Name,
			VRNumber:           vRouter.Number,
			RouteID:            vRouter.RouteID,
			ASNumber:           vRouter.ASNumber,
			PeerASNumber:       vRouter.PeerASNumber,
			RouteDistinguisher: vRouter.RouteDistinguisher,
			Interfaces:         getBGPInterfaces(vRouter.Interfaces),
		}
	}
	return resp
}

func getBGPInterfaces(interfaces []types.NFBGPInterface) []types.BGPInterface {
	resp := make([]types.BGPInterface, len(interfaces))
	for i, interf := range interfaces {
		resp[i] = types.BGPInterface(interf)
	}
	return resp
}

func GetNetworkInterfaces(iProfiles []types.NFTypeInterfaceProfile,
	icMap map[int]types.InterfaceCfgSpec,
	ipMap map[string]*types.InterfaceProfile) ([]types.NetworkInterface, error) {

	resp := make([]types.NetworkInterface, len(iProfiles))
	for i, profile := range iProfiles {
		config, ok := icMap[profile.ID]
		if !ok {
			return nil, fmt.Errorf("did not find interfaceConfig with ID:%d", profile.ID)
		}
		var ip *types.InterfaceProfile = nil
		if ipMap != nil {
			ip, ok = ipMap[profile.InterfaceProfileName]
			if !ok {
				return nil, fmt.Errorf("did not find InterfaceProfile with name:%s", profile.InterfaceProfileName)
			}
		}

		resp[i] = types.NetworkInterface{
			InterfaceName: config.Name,
			IPAddr:        emptyIfNil(config.IPAddr),
			Vlan:          emptyIfNil(config.Vlan),
		}
		if ip != nil {
			resp[i].Latency = ip.Spec.Latency
			resp[i].Bandwidth = ip.Spec.Bandwidth
		}
	}
	return resp, nil
}

func emptyIfNil[T any](slice []T) []T {
	if slice == nil {
		return []T{}
	} else {
		return slice
	}
}
