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

package nftypehydration

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	deployv1alpha1 "github.com/nephio-project/nf-deploy-controller/api/v1alpha1"
	"github.com/nephio-project/nf-deploy-controller/hydration/types"
	"github.com/nephio-project/nf-deploy-controller/hydration/utils"
	ps "github.com/nephio-project/nf-deploy-controller/packageservice"
	nfdeployutil "github.com/nephio-project/nf-deploy-controller/util"
)

const (
	SmfTypeKind            = "SmfType"
	SmfCapacityProfileKind = "SmfCapacityProfile"
	SmfDeployKind          = "SmfDeploy"
	SmfDeployName          = "smfdeploy-%s" // smfdeploy-siteID
)

// SmfDeployImpl implements NfTypeDeployInterface
type SmfDeployImpl struct {
	PS  ps.PackageServiceInterface
	Log logr.Logger
}

// GenerateNfTypeDeploy generates SmfDeploy
func (sdi *SmfDeployImpl) GenerateNfTypeDeploy(
	ctx context.Context,
	s deployv1alpha1.Site, nfDeployName string,
) ([]byte, error) {

	sdi.Log.Info("Generating SmfDeploy", "siteID", s.Id)
	nc, err := nfdeployutil.NewNamingContext(s.ClusterName, nfDeployName)
	if err != nil {
		return nil, fmt.Errorf("error creating naming context: %w", err)
	}
	smfType, err := getSmfType(ctx, sdi.PS, SmfTypeKind, s.NFTypeName, nc)
	if err != nil {
		return nil, fmt.Errorf("error getting SmfType: %w", err)
	}
	nfBgpConfig, interfaceConfigs, err := utils.GetReferencedProfiles(
		ctx, sdi.PS,
		utils.NFBGPConfigKind, utils.InterfaceConfigKind, nc,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting nfProfiles: %w", err)
	}
	cp, err := getSmfCapacityProfile(
		ctx, sdi.PS, SmfCapacityProfileKind,
		smfType.Spec.CapacityProfile.ProfileName, nc,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting SmfCapacityProfile: %w", err)
	}
	// InterfaceConfigMap
	icMap := utils.GetInterfaceConfigSpecMap(interfaceConfigs)
	// InterfaceProfileMap
	ipMap, err := utils.GetInterfaceProfile(
		ctx, sdi.PS, utils.InterfaceProfileKind,
		getInterfaceProfileNames(smfType), nc,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting InterfaceProfile: %w", err)
	}

	smfDeploy, err := generateSmfDeploy(
		s, smfType, cp, nfBgpConfig, icMap, ipMap, nfDeployName,
	)
	if err != nil {
		return nil, fmt.Errorf("error generating SmfDeploy: %w", err)
	}

	content, err := yaml.Marshal(smfDeploy)
	if err != nil {
		return nil, fmt.Errorf("error marshalling the SmfDeploy: %w", err)
	}
	sdi.Log.Info("Generated SmfDeploy Successfully", "siteID", s.Id)
	return content, nil
}

func getInterfaceProfileNames(smfType *types.SmfType) []string {
	ipArr := []types.NFTypeInterfaceProfile{}
	ipArr = append(ipArr, smfType.Spec.N4InterfaceProfile...)
	ipArr = append(ipArr, smfType.Spec.N7InterfaceProfile...)
	ipArr = append(ipArr, smfType.Spec.N10InterfaceProfile...)
	ipArr = append(ipArr, smfType.Spec.N11InterfaceProfile...)
	names := []string{}
	for _, ip := range ipArr {
		names = append(names, ip.InterfaceProfileName)
	}
	return names
}

// generateSmfDeploy generates SmfDeploy
func generateSmfDeploy(
	s deployv1alpha1.Site,
	smfType *types.SmfType,
	cp *types.SmfCapacityProfile,
	nfBgpConfig *types.NFBGPConfig,
	icMap map[int]types.InterfaceCfgSpec,
	ipMap map[string]*types.InterfaceProfile,
	nfDeployName string,
) (*types.SmfDeploy, error) {

	var cap types.SmfCapacity
	if cp != nil {
		cap = types.SmfCapacity{
			MaxSession: cp.MaxSessions,
		}
	}
	n4If, err := utils.GetNetworkInterfaces(
		smfType.Spec.N4InterfaceProfile, icMap, ipMap,
	)
	if err != nil {
		return nil, err
	}
	n7If, err := utils.GetNetworkInterfaces(
		smfType.Spec.N7InterfaceProfile, icMap, ipMap,
	)
	if err != nil {
		return nil, err
	}
	n10If, err := utils.GetNetworkInterfaces(
		smfType.Spec.N10InterfaceProfile, icMap, ipMap,
	)
	if err != nil {
		return nil, err
	}
	n11If, err := utils.GetNetworkInterfaces(
		smfType.Spec.N11InterfaceProfile, icMap, ipMap,
	)
	if err != nil {
		return nil, err
	}
	return &types.SmfDeploy{
		ResourceMeta: yaml.ResourceMeta{
			TypeMeta: yaml.TypeMeta{
				APIVersion: utils.OpAPIVersion,
				Kind:       SmfDeployKind,
			},
			ObjectMeta: yaml.ObjectMeta{
				NameMeta: yaml.NameMeta{
					Name:      fmt.Sprintf(SmfDeployName, s.Id),
					Namespace: utils.OpNameSpace,
				},
				Labels: map[string]string{
					nfdeployutil.NFSiteIDLabel: s.Id,
					nfdeployutil.NFTypeLabel:   s.NFType,
					nfdeployutil.NFDeployLabel: nfDeployName,
				},
			},
		},
		Spec: types.SmfDeploySpec{
			Capacity:      cap,
			BGPConfigs:    utils.GetBGPConfig(nfBgpConfig),
			N4Interfaces:  n4If,
			N7Interfaces:  n7If,
			N10Interfaces: n10If,
			N11Interfaces: n11If,
		},
	}, nil
}

//----------------------------------------------------------
// This section contains packageservice interaction methods
//----------------------------------------------------------

// getSmfType returns the SmfType
func getSmfType(
	ctx context.Context, psi ps.PackageServiceInterface,
	kind, name string, nc nfdeployutil.NamingContext,
) (*types.SmfType, error) {

	nfProfilesMap, err := psi.GetNFProfiles(
		ctx, []ps.GetResourceRequest{
			{
				ID:         1,
				ApiVersion: utils.IpAPIVersion,
				Kind:       kind,
				Name:       name,
			},
		}, nc,
	)
	if err != nil {
		return nil, err
	}
	if len(nfProfilesMap[1]) != 1 {
		return nil, fmt.Errorf(
			"expecting exactly one %s kind with name: %s, received: %d",
			kind, name, len(nfProfilesMap[1]),
		)
	}
	smfType := &types.SmfType{}
	err = yaml.Unmarshal([]byte(nfProfilesMap[1][0]), smfType)
	return smfType, err
}

// getSmfCapacityProfile returns SmfCapacityProfile
func getSmfCapacityProfile(
	ctx context.Context, psi ps.PackageServiceInterface,
	kind, name string, nc nfdeployutil.NamingContext,
) (*types.SmfCapacityProfile, error) {

	cpMap, err := psi.GetNFProfiles(
		ctx, []ps.GetResourceRequest{
			{
				ID:         1,
				ApiVersion: utils.IpAPIVersion,
				Kind:       kind,
				Name:       name,
			},
		}, nc,
	)
	if err != nil {
		return nil, err
	}
	if len(cpMap[1]) != 1 {
		return nil, fmt.Errorf(
			"expecting exactly one %s kind with name: %s, received: %d",
			kind, name, len(cpMap[1]),
		)
	}
	resp := &types.SmfCapacityProfile{}
	err = yaml.Unmarshal([]byte(cpMap[1][0]), resp)
	if err != nil {
		return nil, err
	}
	return resp, err
}
