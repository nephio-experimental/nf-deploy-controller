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
	"errors"

	"github.com/golang/mock/gomock"
	types "github.com/nephio-project/common-lib/nfdeploy"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe(
	"GetSMFIntent", func() {
		var sampleSMFType = "sampleSMFType"
		var expectedSessions = "2000"
		var sampleCapacityProfileName = "capacity-profile"
		var mockManager *MockCRDReader
		BeforeEach(
			func() {
				ctrl := gomock.NewController(GinkgoT())
				mockManager = NewMockCRDReader(ctrl)

			},
		)
		Context(
			"When capacity profile name is given", func() {
				It(
					"Should return expected SMFIntent", func() {
						mockManager.EXPECT().GetSMFTypeObject(
							sampleSMFType,
						).
							Return(
								types.SMFType{
									Spec: types.SMFTypeSpec{CapacityProfile: sampleCapacityProfileName},
								}, nil,
							).AnyTimes()
						mockManager.EXPECT().GetSMFCapacityProfileObject(
							sampleCapacityProfileName,
						).
							Return(
								types.SMFCapacityProfile{
									Spec: types.SMFCapacityProfileSpec{
										MaxSessions: expectedSessions,
									},
								}, nil,
							)
						var smfIntent SMFIntent
						outputIntent, _ := smfIntent.GetSMFIntent(
							sampleSMFType, mockManager,
						)
						Expect(outputIntent.MaxSessions).To(Equal(expectedSessions))
					},
				)
			},
		)
		Context(
			"When SMFType is not present", func() {
				It(
					"Should return error", func() {
						errorString := "SMFType CRD with name " + sampleSMFType + " not found"
						mockManager.EXPECT().GetSMFTypeObject(
							sampleSMFType,
						).Return(
							types.SMFType{}, errors.New(errorString),
						).AnyTimes()
						var smfIntent SMFIntent
						_, err := smfIntent.GetSMFIntent(
							sampleSMFType, mockManager,
						)
						Expect(err).To(Equal(errors.New(errorString)))
					},
				)
			},
		)
		Context(
			"When SMFCapacityProfile is not present", func() {
				It(
					"Should return error", func() {
						errorString := "SMFCapacityProfile CRD with name " + sampleCapacityProfileName + " not found"
						mockManager.EXPECT().GetSMFTypeObject(
							sampleSMFType,
						).
							Return(
								types.SMFType{
									Spec: types.SMFTypeSpec{CapacityProfile: sampleCapacityProfileName},
								}, nil,
							).AnyTimes()
						mockManager.EXPECT().GetSMFCapacityProfileObject(
							sampleCapacityProfileName,
						).Return(
							types.SMFCapacityProfile{}, errors.New(errorString),
						).AnyTimes()
						var smfIntent SMFIntent
						_, err := smfIntent.GetSMFIntent(
							sampleSMFType, mockManager,
						)
						Expect(err).To(Equal(errors.New(errorString)))
					},
				)
			},
		)
	},
)
