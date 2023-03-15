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
	types "github.com/nephio-project/common-lib/nfdeploy"
	crdreader "github.com/nephio-project/nf-deploy-controller/crd-reader"
)

type FakeCRDSet struct {
}

func (fakeCRDSet *FakeCRDSet) ReadCRDFiles(directory string) error {
	return nil
}

func (fakeCRDSet *FakeCRDSet) GetUPFTypeObject(crdName string) (
	types.UPFType, error,
) {
	return types.UPFType{}, nil
}

func (fakeCRDSet *FakeCRDSet) GetSMFTypeObject(crdName string) (
	types.SMFType, error,
) {
	return types.SMFType{}, nil
}

func (fakeCRDSet *FakeCRDSet) GetUPFCapacityProfileObject(crdName string) (
	types.UPFCapacityProfile, error,
) {
	return types.UPFCapacityProfile{}, nil
}

func (fakeCRDSet *FakeCRDSet) GetSMFCapacityProfileObject(crdName string) (
	types.SMFCapacityProfile, error,
) {
	return types.SMFCapacityProfile{}, nil
}

var _ crdreader.CRDReader = &FakeCRDSet{}
