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

package nftypehydration_test

import (
	"context"
	"errors"
	"os"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	ctrl "sigs.k8s.io/controller-runtime"

	deployv1alpha1 "github.com/nephio-project/nf-deploy-controller/api/v1alpha1"
	"github.com/nephio-project/nf-deploy-controller/hydration/nftypehydration"
	hydrationutil "github.com/nephio-project/nf-deploy-controller/hydration/utils"
	ps "github.com/nephio-project/nf-deploy-controller/packageservice"
	mps "github.com/nephio-project/nf-deploy-controller/packageservice/mock"
	nfdeployutil "github.com/nephio-project/nf-deploy-controller/util"
)

const (
	udmNfDeployName = "nfDeploy1"
	udmClusterName  = "cluster1"
)

var (
	udmcp, udmDeploy1 []byte
	udmNC             nfdeployutil.NamingContext
)

func expectUdmCapacityProfile(mpsi *mps.MockPackageServiceInterface) {
	mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
		{
			ID:         1,
			ApiVersion: hydrationutil.IpAPIVersion,
			Kind:       "UdmCapacityProfile",
		},
	}), gomock.Eq(udmNC)).Return(map[int][]string{
		1: {string(udmcp)},
	}, nil).Times(1)
}

var _ = Describe("Udmdeploy", func() {
	var (
		mockCtrl *gomock.Controller
		mpsi     *mps.MockPackageServiceInterface
		udi      *nftypehydration.UdmDeployImpl
	)
	ctx := context.Background()

	udmcp, _ = os.ReadFile("../testhelper/udmcapacityprofile.yaml")
	udmDeploy1, _ = os.ReadFile("../testhelper/udmdeploy1.yaml")
	udmNC, _ = nfdeployutil.NewNamingContext(udmClusterName, udmNfDeployName)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mpsi = mps.NewMockPackageServiceInterface(mockCtrl)
		udi = &nftypehydration.UdmDeployImpl{
			PS:  mpsi,
			Log: ctrl.Log.WithName("udmdeploy"),
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Testing GenerateNfTypeDeploy for single udm site", func() {
		site := deployv1alpha1.Site{
			Id:          "udm1",
			ClusterName: udmClusterName,
			NFType:      "udm",
			NFTypeName:  "udmsmall",
			NFVendor:    "casa",
			NFVersion:   "1.0",
		}
		Context("testing udmdeploy", func() {
			BeforeEach(func() {
				expectUdmCapacityProfile(mpsi)
			})
			It("should process a single udm and return udmdeploy", func() {
				format.MaxLength = 0
				resp, err := udi.GenerateNfTypeDeploy(ctx, site, udmNfDeployName)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).To(Equal(udmDeploy1))
			})
		})
		Context("expecting no value from packageservice for GetNFProfiles", func() {
			BeforeEach(func() {
				mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
					{
						ID:         1,
						ApiVersion: hydrationutil.IpAPIVersion,
						Kind:       "UdmCapacityProfile",
					},
				}), gomock.Eq(udmNC)).Return(map[int][]string{
					1: {},
				}, nil).Times(1)
			})
			It("should process a single udm and return error", func() {
				resp, err := udi.GenerateNfTypeDeploy(ctx, site, udmNfDeployName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("error getting UdmCapacityProfile"))
				Expect(resp).To(BeNil())
			})
		})
		Context("expecting error from packageservice for GetNFProfiles", func() {
			expectedErr := errors.New("error from packageservice")
			It("should return an error while getting udmCapacityProfile from packageservice", func() {
				mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
					{
						ID:         1,
						ApiVersion: hydrationutil.IpAPIVersion,
						Kind:       "UdmCapacityProfile",
					},
				}), gomock.Eq(udmNC)).Return(nil, expectedErr).Times(1)

				resp, err := udi.GenerateNfTypeDeploy(ctx, site, udmNfDeployName)
				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())
				Expect(err.Error()).To(HaveSuffix(expectedErr.Error()))
			})
		})
	})
})
