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

	udmtypes "github.com/nephio-project/common-lib/udm"
	deployv1alpha1 "github.com/nephio-project/nf-deploy-controller/api/v1alpha1"
	"github.com/nephio-project/nf-deploy-controller/hydration/types"
	"github.com/nephio-project/nf-deploy-controller/hydration/utils"
	ps "github.com/nephio-project/nf-deploy-controller/packageservice"
	nfdeployutil "github.com/nephio-project/nf-deploy-controller/util"
)

const (
	UdmCapacityProfileKind = "UdmCapacityProfile"
	UdmDeployKind          = "UdmDeploy"
	UdmDeployName          = "udmdeploy-%s" // udmdeploy-siteID
)

// UdmDeployImpl implements NfTypeDeployInterface
type UdmDeployImpl struct {
	PS  ps.PackageServiceInterface
	Log logr.Logger
}

// GenerateNfTypeDeploy generates UdmDeploy
func (udi *UdmDeployImpl) GenerateNfTypeDeploy(
	ctx context.Context,
	s deployv1alpha1.Site, nfDeployName string,
) ([]byte, error) {

	udi.Log.Info("Generating UdmDeploy", "siteID", s.Id)
	nc, err := nfdeployutil.NewNamingContext(s.ClusterName, nfDeployName)
	if err != nil {
		return nil, fmt.Errorf("error creating naming context: %w", err)
	}
	cp, err := getUdmCapacityProfile(ctx, udi.PS, UdmCapacityProfileKind, nc)
	if err != nil {
		return nil, fmt.Errorf("error getting UdmCapacityProfile: %w", err)
	}

	udmDeploy := generateUdmDeploy(s, cp, nfDeployName)

	content, err := yaml.Marshal(udmDeploy)
	if err != nil {
		return nil, fmt.Errorf("error marshalling the UdmDeploy: %w", err)
	}
	udi.Log.Info("Generated UdmDeploy Successfully", "siteID", s.Id)
	return content, nil
}

// generateUdmDeploy generates UdmDeploy
func generateUdmDeploy(
	s deployv1alpha1.Site,
	cp *types.UdmCapacityProfile,
	nfDeployName string,
) *types.UdmDeploy {

	return &types.UdmDeploy{
		ResourceMeta: yaml.ResourceMeta{
			TypeMeta: yaml.TypeMeta{
				APIVersion: utils.OpAPIVersion,
				Kind:       UdmDeployKind,
			},
			ObjectMeta: yaml.ObjectMeta{
				NameMeta: yaml.NameMeta{
					Name:      fmt.Sprintf(UdmDeployName, s.Id),
					Namespace: utils.OpNameSpace,
				},
				Labels: map[string]string{
					nfdeployutil.NFSiteIDLabel: s.Id,
					nfdeployutil.NFTypeLabel:   s.NFType,
					nfdeployutil.NFDeployLabel: nfDeployName,
				},
			},
		},
		Spec: udmtypes.UdmDeploySpec{
			CapacityProfile: udmtypes.CapacityProfile{
				RequestedCpu:    cp.Spec.RequestedCpu,
				RequestedMemory: cp.Spec.RequestedMemory,
			},
			NfInfo: udmtypes.NfInfo{
				Vendor:  s.NFVendor,
				Version: s.NFVersion,
			},
		},
	}
}

//----------------------------------------------------------
// This section contains packageservice interaction methods
//----------------------------------------------------------

// getUdmCapacityProfile returns UdmCapacityProfile
func getUdmCapacityProfile(
	ctx context.Context, psi ps.PackageServiceInterface,
	kind string, nc nfdeployutil.NamingContext,
) (*types.UdmCapacityProfile, error) {

	cpMap, err := psi.GetNFProfiles(
		ctx, []ps.GetResourceRequest{
			{
				ID:         1,
				ApiVersion: utils.IpAPIVersion,
				Kind:       kind,
			},
		}, nc,
	)
	if err != nil {
		return nil, err
	}
	if len(cpMap[1]) != 1 {
		return nil, fmt.Errorf(
			"expecting exactly one %s kind, received: %d",
			kind, len(cpMap[1]),
		)
	}
	resp := &types.UdmCapacityProfile{}
	err = yaml.Unmarshal([]byte(cpMap[1][0]), resp)
	if err != nil {
		return nil, err
	}
	return resp, err
}
