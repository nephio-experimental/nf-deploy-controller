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

package crdreader

import (
	types "github.com/nephio-project/common-lib/nfdeploy"
)

// UPFIntentProcessor : UPFIntentProcessor is an interface which extracts
// upf specs and intent from CRDs specified in NFDeploy
type UPFIntentProcessor interface {

	// getUPFCapacityProfileName : returns UPF capacity profile name from upf Type file
	getUPFCapacityProfileName(upfTypeName string, crdReader CRDReader) (
		string, error,
	)

	// getUPFCapacityProfile : returns UPFCapacityProfile struct from capacity profile name
	getUPFCapacityProfile(
		capacityProfileName string, crdReader CRDReader,
	) (types.UPFCapacityProfile, error)

	// GetUPFIntent : returns intent required from UPF using upfType Name
	GetUPFIntent(upfTypeName string, crdReader CRDReader) (UPFIntent, error)
}
