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
	. "github.com/nephio-project/common-lib/nfdeploy"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe(
	"GetUPFIntent", func() {
		var sampleUPFType = "sampleUPFType"
		var expectedThroughput = "2000"
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
					"Should return expected UPFIntent", func() {
						mockManager.EXPECT().GetUPFTypeObject(
							sampleUPFType,
						).
							Return(
								UPFType{
									Spec: UPFTypeSpec{CapacityProfile: sampleCapacityProfileName},
								}, nil,
							).AnyTimes()
						mockManager.EXPECT().GetUPFCapacityProfileObject(
							sampleCapacityProfileName,
						).
							Return(
								UPFCapacityProfile{
									Spec: UPFCapacityProfileSpec{
										Throughput: expectedThroughput,
									},
								}, nil,
							)
						var upfIntent UPFIntent
						outputIntent, _ := upfIntent.GetUPFIntent(
							sampleUPFType, mockManager,
						)
						Expect(outputIntent.Throughput).To(Equal(expectedThroughput))
					},
				)
			},
		)
		Context(
			"When UPFType is not present", func() {
				It(
					"Should return error", func() {
						errorString := "UPFType CRD with name " + sampleUPFType + " not found"
						mockManager.EXPECT().GetUPFTypeObject(
							sampleUPFType,
						).Return(
							UPFType{}, errors.New(errorString),
						).AnyTimes()
						var upfIntent UPFIntent
						_, err := upfIntent.GetUPFIntent(
							sampleUPFType, mockManager,
						)
						Expect(err).To(Equal(errors.New(errorString)))
					},
				)
			},
		)
		Context(
			"When UPFCapacityProfile is not present", func() {
				It(
					"Should return error", func() {
						errorString := "UPFCapacityProfile CRD with name " + sampleCapacityProfileName + " not found"
						mockManager.EXPECT().GetUPFTypeObject(
							sampleUPFType,
						).
							Return(
								UPFType{
									Spec: UPFTypeSpec{CapacityProfile: sampleCapacityProfileName},
								}, nil,
							).AnyTimes()
						mockManager.EXPECT().GetUPFCapacityProfileObject(
							sampleCapacityProfileName,
						).Return(
							UPFCapacityProfile{}, errors.New(errorString),
						).AnyTimes()
						var upfIntent UPFIntent
						_, err := upfIntent.GetUPFIntent(
							sampleUPFType, mockManager,
						)
						Expect(err).To(Equal(errors.New(errorString)))
					},
				)
			},
		)
	},
)
