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

	deployv1alpha1 "github.com/nephio-project/nf-deploy-controller/api/v1alpha1"
)

type NfTypeHydrationInterface interface {
	// GenerateNfTypeDeploy generates NfTypeDeploy (UpfDeploy, SmfDeploy)
	// and returns the generates resource as []byte
	GenerateNfTypeDeploy(ctx context.Context, s deployv1alpha1.Site, nfDeployName string) ([]byte, error)
}
