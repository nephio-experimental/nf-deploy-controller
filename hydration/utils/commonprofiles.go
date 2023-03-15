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
	"context"
	"fmt"

	"sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/nephio-seed/nf-deploy-controller/hydration/types"
	ps "github.com/nephio-seed/nf-deploy-controller/packageservice"
	nfdeployutil "github.com/nephio-seed/nf-deploy-controller/util"
)

//===========================================================================
// This file contains generic methods which can be used by
// all the NfTypes (UPF, SMF) to interact with packageservice
//===========================================================================

// GetReferencedProfiles returns the nfBgpConfig, interfaceConfig
func GetReferencedProfiles(ctx context.Context, psi ps.PackageServiceInterface, nfBgpKind,
	interfaceConfigKind string, nc nfdeployutil.NamingContext) (*types.NFBGPConfig, []*types.InterfaceConfig, error) {

	nfProfilesMap, err := psi.GetNFProfiles(ctx, []ps.GetResourceRequest{
		{
			ID:         1,
			ApiVersion: IpAPIVersion,
			Kind:       nfBgpKind,
		},
		{
			ID:         2,
			ApiVersion: IpAPIVersion,
			Kind:       interfaceConfigKind,
		},
	}, nc)
	if err != nil {
		return nil, nil, err
	}
	// Assumption: 	We are assuming that there will be at the most one NfBgpConfig file containing details
	//   about all the virtualRouters as there is no name for NfBgpConfig specified in NfDeploy/NfType CRDs
	//   to fetch the specific NfBgpConfig.
	if len(nfProfilesMap[1]) > 1 {
		return nil, nil, fmt.Errorf("expecting at the most one %s kind, received: %d",
			nfBgpKind, len(nfProfilesMap[1]))
	}
	nfBgpConfig := &types.NFBGPConfig{}
	if len(nfProfilesMap[1]) == 0 {
		nfBgpConfig = nil
	} else {
		err = yaml.Unmarshal([]byte(nfProfilesMap[1][0]), nfBgpConfig)
		if err != nil {
			return nil, nil, err
		}
	}
	ic := make([]*types.InterfaceConfig, len(nfProfilesMap[2]))
	for i := range nfProfilesMap[2] {
		ic[i] = &types.InterfaceConfig{}
		err = yaml.Unmarshal([]byte(nfProfilesMap[2][i]), ic[i])
		if err != nil {
			return nil, nil, err
		}
	}
	return nfBgpConfig, ic, err
}

// GetInterfaceProfile returns InterfaceProfile map with name as key
func GetInterfaceProfile(ctx context.Context, psi ps.PackageServiceInterface,
	kind string, names []string, nc nfdeployutil.NamingContext) (map[string]*types.InterfaceProfile, error) {

	nameIDMap := make(map[string]int) // this map is used for deduplication
	idNameMap := make(map[int]string)
	req := []ps.GetResourceRequest{}
	for i, n := range names {
		_, ok := nameIDMap[n]
		if !ok {
			nameIDMap[n] = i
			idNameMap[i] = n
			req = append(req, ps.GetResourceRequest{
				ID:         i,
				ApiVersion: IpAPIVersion,
				Kind:       kind,
				Name:       n,
			})
		}
	}
	cpMap, err := psi.GetNFProfiles(ctx, req, nc)
	if err != nil {
		return nil, err
	}
	resp := make(map[string]*types.InterfaceProfile)
	for id := range cpMap {
		ip := &types.InterfaceProfile{}
		err = yaml.Unmarshal([]byte(cpMap[id][0]), ip)
		if err != nil {
			return nil, err
		}
		resp[idNameMap[id]] = ip
	}
	return resp, nil
}
