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

type SMFIntent struct {
	MaxSessions string
}

var _ SMFIntentProcessor = &SMFIntent{}

// getSMFCapacityProfileName : returns SMF capacity profile name from smf Type file
func (smfIntent *SMFIntent) getSMFCapacityProfileName(
	smfTypeName string, crdReader CRDReader,
) (string, error) {
	smfType, err := crdReader.GetSMFTypeObject(smfTypeName)
	if err != nil {
		return "", err
	}
	return smfType.Spec.CapacityProfile, nil
}

// getSMFCapacityProfile : returns SMFCapacityProfile struct from capacity profile name
func (smfIntent *SMFIntent) getSMFCapacityProfile(
	capacityProfileName string, crdReader CRDReader,
) (SMFCapacityProfile, error) {
	smfCapacityProfile, err := crdReader.GetSMFCapacityProfileObject(
		capacityProfileName,
	)
	if err != nil {
		return SMFCapacityProfile{}, err
	}
	return smfCapacityProfile, nil
}

// GetSMFIntent : returns intent required from SMF using smfType Name
func (smfIntent *SMFIntent) GetSMFIntent(
	smfTypeName string, crdReader CRDReader,
) (SMFIntent, error) {
	capacityProfileName, err := smfIntent.getSMFCapacityProfileName(
		smfTypeName, crdReader,
	)
	if err != nil {
		return SMFIntent{}, err
	}
	capacityProfile, err := smfIntent.getSMFCapacityProfile(
		capacityProfileName, crdReader,
	)
	if err != nil {
		return SMFIntent{}, err
	}
	return SMFIntent{
		MaxSessions: capacityProfile.Spec.MaxSessions,
	}, nil
}
