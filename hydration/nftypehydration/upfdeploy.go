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
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	yamlutil "sigs.k8s.io/yaml"

	deployv1alpha1 "github.com/nephio-project/nf-deploy-controller/api/v1alpha1"
	"github.com/nephio-project/nf-deploy-controller/hydration/types"
	"github.com/nephio-project/nf-deploy-controller/hydration/utils"
	ps "github.com/nephio-project/nf-deploy-controller/packageservice"
	nfdeployutil "github.com/nephio-project/nf-deploy-controller/util"
)

const (
	UpfTypeKind            = "UpfType"
	UpfCapacityProfileKind = "UpfCapacityProfile"
	UpfDeployKind          = "UpfDeploy"
	UpfDeployName          = "upfdeploy-%s"           // upfdeploy-siteID
	UpfDeployExtensionName = "upfdeploy-%s-extension" // upfdeploy-siteID
)

// UpfDeployImpl implements NfTypeDeployInterface
type UpfDeployImpl struct {
	PS  ps.PackageServiceInterface
	Log logr.Logger
}

// GenerateNfTypeDeploy generates UpfDeploy
func (udi *UpfDeployImpl) GenerateNfTypeDeploy(
	ctx context.Context,
	s deployv1alpha1.Site, nfDeployName string,
) ([]byte, error) {

	udi.Log.Info("Generating UpfDeploy", "siteID", s.Id)
	nc, err := nfdeployutil.NewNamingContext(s.ClusterName, nfDeployName)
	if err != nil {
		return nil, fmt.Errorf("error creating naming context: %w", err)
	}
	upfType, err := getUpfType(ctx, udi.PS, UpfTypeKind, s.NFTypeName, nc)
	if err != nil {
		return nil, fmt.Errorf("error getting UpfType: %w", err)
	}
	nfBgpConfig, interfaceConfigs, err := utils.GetReferencedProfiles(
		ctx, udi.PS,
		utils.NFBGPConfigKind, utils.InterfaceConfigKind, nc,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting nfProfiles: %w", err)
	}
	cp, err := getUpfCapacityProfile(
		ctx, udi.PS, UpfCapacityProfileKind,
		upfType.Spec.UpfCapacityProfile.ProfileName, nc,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting UpfCapacityProfile: %w", err)
	}
	icMap := utils.GetInterfaceConfigSpecMap(interfaceConfigs)

	upfDeploy, err := generateUpfDeploy(
		s, upfType, cp, nfBgpConfig, icMap, nfDeployName,
	)
	if err != nil {
		return nil, fmt.Errorf("error generating UpfDeploy: %w", err)
	}

	extnString, extnObj, err := udi.getUpfVendorExtnObj(ctx, nc, s)
	if err != nil {
		return nil, fmt.Errorf("error generating UpfDeploy: %w", err)
	}
	upfDeploy.Spec.VendorRef = extnObj

	content, err := yamlutil.Marshal(upfDeploy)
	if err != nil {
		return nil, fmt.Errorf("error marshalling the UpfDeploy: %w", err)
	}
	if extnObj != nil {
		content = []byte(string(content) + nfdeployutil.YamlObjectDelimiter + "\n" + extnString)
	}
	udi.Log.Info("Generated UpfDeploy Successfully", "siteID", s.Id)
	return content, nil
}

// generateUpfDeploy generates UpfDeploy
func generateUpfDeploy(
	s deployv1alpha1.Site,
	upfType *types.UpfType,
	cp *types.UpfCapacityProfile,
	nfBgpConfig *types.NFBGPConfig,
	icMap map[int]types.InterfaceCfgSpec,
	nfDeployName string,
) (*types.UpfDeploy, error) {

	var cap types.UpfCapacity
	if cp != nil {
		cap = cp.UpfCPSpec.UpfCapacity
	}
	n3If, err := utils.GetNetworkInterfaces(
		upfType.Spec.N3InterfaceProfile, icMap, nil,
	)
	if err != nil {
		return nil, err
	}
	n4If, err := utils.GetNetworkInterfaces(
		upfType.Spec.N4InterfaceProfile, icMap, nil,
	)
	if err != nil {
		return nil, err
	}
	n6If, err := utils.GetNetworkInterfaces(
		upfType.Spec.N6InterfaceProfile, icMap, nil,
	)
	if err != nil {
		return nil, err
	}
	n9If, err := utils.GetNetworkInterfaces(
		upfType.Spec.N9InterfaceProfile, icMap, nil,
	)
	if err != nil {
		return nil, err
	}
	return &types.UpfDeploy{
		ResourceMeta: yaml.ResourceMeta{
			TypeMeta: yaml.TypeMeta{
				APIVersion: utils.OpAPIVersion,
				Kind:       UpfDeployKind,
			},
			ObjectMeta: yaml.ObjectMeta{
				NameMeta: yaml.NameMeta{
					Name:      fmt.Sprintf(UpfDeployName, s.Id),
					Namespace: utils.OpNameSpace,
				},
				Labels: map[string]string{
					nfdeployutil.NFSiteIDLabel: s.Id,
					nfdeployutil.NFTypeLabel:   s.NFType,
					nfdeployutil.NFDeployLabel: nfDeployName,
				},
			},
		},
		Spec: types.UpfDeploySpec{
			Capacity:     cap,
			BGPConfigs:   utils.GetBGPConfig(nfBgpConfig),
			N3Interfaces: n3If,
			N4Interfaces: n4If,
			N6Interfaces: n6If,
			N9Interfaces: n9If,
		},
	}, nil
}

//----------------------------------------------------------
// This section contains packageservice interaction methods
//----------------------------------------------------------

// getUpfType returns the UpfType
func getUpfType(
	ctx context.Context, psi ps.PackageServiceInterface,
	kind, name string, nc nfdeployutil.NamingContext,
) (*types.UpfType, error) {

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
	upfType := &types.UpfType{}
	err = yamlutil.Unmarshal([]byte(nfProfilesMap[1][0]), upfType)
	return upfType, err
}

// getUpfCapacityProfile returns UpfCapacityProfile
func getUpfCapacityProfile(
	ctx context.Context, psi ps.PackageServiceInterface,
	kind, name string, nc nfdeployutil.NamingContext,
) (*types.UpfCapacityProfile, error) {

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
	resp := &types.UpfCapacityProfile{}
	err = yamlutil.Unmarshal([]byte(cpMap[1][0]), resp)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func (udi *UpfDeployImpl) getUpfVendorExtnObj(
	ctx context.Context,
	nc nfdeployutil.NamingContext,
	s deployv1alpha1.Site,
) (string, *types.ObjectReference, error) {
	key := ps.VendorNFKey{
		Vendor:  s.NFVendor,
		Version: s.NFVersion,
		NFType:  s.NFType,
	}

	extnObjs, err := udi.PS.GetVendorExtensionPackage(ctx, nc, key)
	if err != nil {
		return "", nil, err
	} else if len(extnObjs) > 1 {
		return "", nil, errors.New(
			fmt.Sprintf(
				"More than one extension object found for vendor nf %#v", key,
			),
		)
	} else if len(extnObjs) == 0 {
		return "", nil, nil
	} else {
		rNodes, err := nfdeployutil.ParseStringToYamlNode(extnObjs[0])
		if err != nil {
			return "", nil, err
		}
		rNodes[0].SetName(fmt.Sprintf(UpfDeployExtensionName, s.Id))
		objRef := types.ObjectReference{
			APIGroup:  rNodes[0].GetApiVersion(),
			Kind:      rNodes[0].GetKind(),
			Namespace: rNodes[0].GetNamespace(),
			Name:      rNodes[0].GetName(),
		}
		return rNodes[0].MustString(), &objRef, nil
	}
}
