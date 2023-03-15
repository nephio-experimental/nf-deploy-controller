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
	. "github.com/nephio-project/common-lib/nfdeploy"
)

type UPFIntent struct {
	Throughput string
}

var _ UPFIntentProcessor = &UPFIntent{}

// getUPFCapacityProfileName : returns UPF capacity profile name from upf Type file
func (upfIntent *UPFIntent) getUPFCapacityProfileName(
	upfTypeName string, crdReader CRDReader,
) (string, error) {
	upfType, err := crdReader.GetUPFTypeObject(upfTypeName)
	if err != nil {
		return "", err
	}
	return upfType.Spec.CapacityProfile, nil
}

// getUPFCapacityProfile : returns UPFCapacityProfile struct from capacity profile name
func (upfIntent *UPFIntent) getUPFCapacityProfile(
	capacityProfileName string, crdReader CRDReader,
) (UPFCapacityProfile, error) {
	upfCapacityProfile, err := crdReader.GetUPFCapacityProfileObject(capacityProfileName)
	if err != nil {
		return UPFCapacityProfile{}, err
	}
	return upfCapacityProfile, nil
}

// GetUPFIntent : returns intent required from UPF using upfType Name
func (upfIntent *UPFIntent) GetUPFIntent(
	upfTypeName string, crdReader CRDReader,
) (UPFIntent, error) {
	capacityProfileName, err := upfIntent.getUPFCapacityProfileName(
		upfTypeName, crdReader,
	)
	if err != nil {
		return UPFIntent{}, err
	}
	capacityProfile, err := upfIntent.getUPFCapacityProfile(
		capacityProfileName, crdReader,
	)
	if err != nil {
		return UPFIntent{}, err
	}
	return UPFIntent{
		Throughput: capacityProfile.Spec.Throughput,
	}, nil
}
