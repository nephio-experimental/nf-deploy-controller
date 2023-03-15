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

	ausftypes "github.com/nephio-project/common-lib/ausf"
	deployv1alpha1 "github.com/nephio-project/nf-deploy-controller/api/v1alpha1"
	"github.com/nephio-project/nf-deploy-controller/hydration/types"
	"github.com/nephio-project/nf-deploy-controller/hydration/utils"
	ps "github.com/nephio-project/nf-deploy-controller/packageservice"
	nfdeployutil "github.com/nephio-project/nf-deploy-controller/util"
)

const (
	AusfCapacityProfileKind = "AusfCapacityProfile"
	AusfDeployKind          = "AusfDeploy"
	AusfDeployName          = "ausfdeploy-%s" // ausfdeploy-siteID
)

// AusfDeployImpl implements NfTypeDeployInterface
type AusfDeployImpl struct {
	PS  ps.PackageServiceInterface
	Log logr.Logger
}

// GenerateNfTypeDeploy generates AusfDeploy
func (adi *AusfDeployImpl) GenerateNfTypeDeploy(
	ctx context.Context,
	s deployv1alpha1.Site, nfDeployName string,
) ([]byte, error) {

	adi.Log.Info("Generating AusfDeploy", "siteID", s.Id)
	nc, err := nfdeployutil.NewNamingContext(s.ClusterName, nfDeployName)
	if err != nil {
		return nil, fmt.Errorf("error creating naming context: %w", err)
	}
	cp, err := getAusfCapacityProfile(ctx, adi.PS, AusfCapacityProfileKind, nc)
	if err != nil {
		return nil, fmt.Errorf("error getting AusfCapacityProfile: %w", err)
	}
	ausfDeploy := generateAusfDeploy(s, cp, nfDeployName)

	content, err := yaml.Marshal(ausfDeploy)
	if err != nil {
		return nil, fmt.Errorf("error marshalling the AusfDeploy: %w", err)
	}
	adi.Log.Info("Generated AusfDeploy Successfully", "siteID", s.Id)
	return content, nil
}

// generateAusfDeploy generates AusfDeploy
func generateAusfDeploy(
	s deployv1alpha1.Site,
	cp *types.AusfCapacityProfile,
	nfDeployName string,
) *types.AusfDeploy {

	return &types.AusfDeploy{
		ResourceMeta: yaml.ResourceMeta{
			TypeMeta: yaml.TypeMeta{
				APIVersion: utils.OpAPIVersion,
				Kind:       AusfDeployKind,
			},
			ObjectMeta: yaml.ObjectMeta{
				NameMeta: yaml.NameMeta{
					Name:      fmt.Sprintf(AusfDeployName, s.Id),
					Namespace: utils.OpNameSpace,
				},
				Labels: map[string]string{
					nfdeployutil.NFSiteIDLabel: s.Id,
					nfdeployutil.NFTypeLabel:   s.NFType,
					nfdeployutil.NFDeployLabel: nfDeployName,
				},
			},
		},
		Spec: ausftypes.AusfDeploySpec{
			CapacityProfile: ausftypes.CapacityProfile{
				RequestedCpu:    cp.Spec.RequestedCpu,
				RequestedMemory: cp.Spec.RequestedMemory,
			},
			NfInfo: ausftypes.NfInfo{
				Vendor:  s.NFVendor,
				Version: s.NFVersion,
			},
		},
	}
}

//----------------------------------------------------------
// This section contains packageservice interaction methods
//----------------------------------------------------------

// getAusfCapacityProfile returns AusfCapacityProfile
func getAusfCapacityProfile(
	ctx context.Context, psi ps.PackageServiceInterface,
	kind string, nc nfdeployutil.NamingContext,
) (*types.AusfCapacityProfile, error) {

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
	resp := &types.AusfCapacityProfile{}
	err = yaml.Unmarshal([]byte(cpMap[1][0]), resp)
	if err != nil {
		return nil, err
	}
	return resp, err
}
