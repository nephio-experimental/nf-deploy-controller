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
	"errors"

	deployv1alpha1 "github.com/nephio-project/nf-deploy-controller/api/v1alpha1"
	"github.com/nephio-project/nf-deploy-controller/hydration"
)

type FakeHydration struct {
}

func (fakeHydration *FakeHydration) Hydrate(
	ctx context.Context, nfDeploy deployv1alpha1.NfDeploy,
) ([]string, error) {
	if nfDeploy.Name == "hydration-failed" {
		return nil, errors.New("error from porch")
	}
	return []string{"resourceName"}, nil
}

func (fakeHydration *FakeHydration) CreateNFDeployActuators(
	ctx context.Context, nfDeploy deployv1alpha1.NfDeploy,
) ([]string, error) {
	if nfDeploy.Name == "actuation-failed" {
		return nil, errors.New("error from porch")
	}
	return []string{"operator-resourceName"}, nil
}

var _ hydration.HydrationInterface = &FakeHydration{}
