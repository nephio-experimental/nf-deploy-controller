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

// SMFIntentProcessor : SMFIntentProcessor is an interface which extracts
// smf specs and intent from CRDs specified in NFDeploy
type SMFIntentProcessor interface {

	// getSMFCapacityProfileName : returns SMF capacity profile name from smf Type file
	getSMFCapacityProfileName(
		smfTypeName string, crdReader CRDReader,
	) (string, error)

	// getSMFCapacityProfile : returns SMFCapacityProfile struct from capacity profile name
	getSMFCapacityProfile(
		capacityProfileName string,
		crdReader CRDReader,
	) (types.SMFCapacityProfile, error)

	// GetSMFIntent : returns intent required from SMF using smfType Name
	GetSMFIntent(
		smfTypeName string, crdReader CRDReader,
	) (SMFIntent, error)
}
