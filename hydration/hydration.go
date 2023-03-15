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

package hydration

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	deployv1alpha1 "github.com/nephio-project/nf-deploy-controller/api/v1alpha1"
	"github.com/nephio-project/nf-deploy-controller/hydration/nftypehydration"
	"github.com/nephio-project/nf-deploy-controller/hydration/utils"
	ps "github.com/nephio-project/nf-deploy-controller/packageservice"
	nfdeployutil "github.com/nephio-project/nf-deploy-controller/util"
)

var (
	upfHydration, smfHydration, ausfHydration, udmHydration nftypehydration.NfTypeHydrationInterface
)

// HydrationInterface is the interface that wraps the steps involved in
// hydration of a NFDeploy into the individual NFType deploys and all other
// supporting manifests like operators required to meet the intent of NFDeploy.
type HydrationInterface interface {
	Hydrate(ctx context.Context, nfDeploy deployv1alpha1.NfDeploy) ([]string, error)
	CreateNFDeployActuators(ctx context.Context, nfDeploy deployv1alpha1.NfDeploy) ([]string, error)
}

type Hydration struct {
	PS  ps.PackageServiceInterface
	Log logr.Logger
}

func (h *Hydration) initHydration() {
	upfHydration = &nftypehydration.UpfDeployImpl{
		PS:  h.PS,
		Log: h.Log,
	}
	smfHydration = &nftypehydration.SmfDeployImpl{
		PS:  h.PS,
		Log: h.Log,
	}
	ausfHydration = &nftypehydration.AusfDeployImpl{
		PS:  h.PS,
		Log: h.Log,
	}
	udmHydration = &nftypehydration.UdmDeployImpl{
		PS:  h.PS,
		Log: h.Log,
	}
}

// Hydrate hydrates the given nfDeploy and generates the NfTypeDeploy (like UpfDeploy, SmfDeploy)
// and creates the packages of generated artifacts using packageservice
func (h *Hydration) Hydrate(ctx context.Context, nfDeploy deployv1alpha1.NfDeploy) ([]string, error) {
	h.Log.Info("Starting Hydration", "nfDeployName", nfDeploy.Name)
	h.initHydration()
	packageContents := make(map[string]map[string]string)
	errSiteIDs := []string{}
	for _, s := range nfDeploy.Spec.Sites {
		h.Log.Info("Processing site", "nfDeployName", nfDeploy.Name, "siteID", s.Id)
		content, err := h.processSite(ctx, s, nfDeploy.Name)
		if err != nil {
			// We are logging the actual error here as only siteIDs are returned to parent function
			h.Log.Error(err, "Error processing site", "nfDeployName", nfDeploy.Name, "siteID", s.Id)
			errSiteIDs = append(errSiteIDs, s.Id)
			continue
		}
		m, ok := packageContents[s.ClusterName]
		if !ok {
			m = make(map[string]string)
		}
		m[fmt.Sprintf(utils.OpFileName, nfDeploy.Name, s.Id)] = string(content)
		packageContents[s.ClusterName] = m
		h.Log.Info("Processed site successfully", "nfDeployName", nfDeploy.Name, "siteID", s.Id)
	}
	if len(errSiteIDs) > 0 {
		return nil, fmt.Errorf("error hydrating sites: %v", errSiteIDs)
	}
	names := []string{}
	for cluster, val := range packageContents {
		nc, err := nfdeployutil.NewNamingContext(cluster, nfDeploy.Name)
		if err != nil {
			return nil, fmt.Errorf("error creating naming context: %w", err)
		}
		n, err := h.PS.CreateDeployPackage(ctx, val, nc)
		if err != nil {
			return nil, fmt.Errorf("error creating package for cluster: %s, err: %w", cluster, err)
		}
		names = append(names, n)
		h.Log.Info("Created porch package", "name", n, "nfDeployName", nfDeploy.Name)
	}
	h.Log.Info("Hydration Successful", "nfDeployName", nfDeploy.Name)
	return names, nil
}

// For each unique vendor, version and nfType in the NFDeploy, creates
// a package in the respective edge deploy repo of the cluster. The package
// contains the operators required for the NFDeploy to be actuated on the
// edge.
// It ensures that these operators are deployed only once on the edge.
// It returns the list of package names to be added for approval if any.
func (h *Hydration) CreateNFDeployActuators(ctx context.Context, nfDeploy deployv1alpha1.NfDeploy) ([]string, error) {
	pkgNamesForApproval := []string{}
	clusterVendorNFsMap := map[string]map[ps.VendorNFKey]bool{}
	for _, site := range nfDeploy.Spec.Sites {
		key := ps.VendorNFKey{
			Vendor:  site.NFVendor,
			Version: site.NFVersion,
			NFType:  site.NFType,
		}
		_, ok := clusterVendorNFsMap[site.ClusterName]
		if !ok {
			clusterVendorNFsMap[site.ClusterName] = map[ps.VendorNFKey]bool{}
		}
		clusterVendorNFsMap[site.ClusterName][key] = true
	}

	for cluster, vendorNFs := range clusterVendorNFsMap {
		nc, err := nfdeployutil.NewNamingContext(cluster, nfDeploy.Name)
		if err != nil {
			return nil, fmt.Errorf("Error creating actuators, error creating naming context: %w", err)
		}
		for vendorNF, _ := range vendorNFs {
			h.Log.V(1).Info(fmt.Sprintf("Creating NFDeployActuators for %s NFDeploy with key:%#v",
				nfDeploy.Name, vendorNF))
			pkgName, isNew, err := h.PS.CreateNFDeployActuators(ctx, nc, vendorNF)
			if err != nil {
				return nil, fmt.Errorf("Error creating actuators with key:%#v , %w", vendorNF, err)
			}
			if !isNew {
				h.Log.V(1).Info(fmt.Sprintf("NFDeployActuators for %s NFDeploy with key:%#v already present: %s",
					nfDeploy.Name, vendorNF, pkgName))
			} else {
				h.Log.V(1).Info(
					fmt.Sprintf("Successfully created NFDeployActuators for %s NFDeploy with key:%#v, package name: %s",
						nfDeploy.Name, vendorNF, pkgName))
				pkgNamesForApproval = append(pkgNamesForApproval, pkgName)
			}
		}
	}
	return pkgNamesForApproval, nil
}

// processSite processes each site from nfDeploy
func (h *Hydration) processSite(ctx context.Context, s deployv1alpha1.Site, nfDeployName string) ([]byte, error) {
	var nfHydration nftypehydration.NfTypeHydrationInterface
	switch s.NFType {
	case utils.UPFKind:
		nfHydration = upfHydration
	case utils.SMFKind:
		nfHydration = smfHydration
	case utils.AUSFKind:
		nfHydration = ausfHydration
	case utils.UDMKind:
		nfHydration = udmHydration
	default:
		return nil, fmt.Errorf("invalid NfType:%s", s.NFType)
	}
	content, err := nfHydration.GenerateNfTypeDeploy(ctx, s, nfDeployName)
	if err != nil {
		return nil, fmt.Errorf("error generating nftypedeploy: %w", err)
	}
	return content, nil
}
