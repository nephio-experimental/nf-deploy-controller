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
	ausfNfDeployName = "nfDeploy1"
	ausfClusterName  = "cluster1"
)

var (
	ausfcp, ausfDeploy1 []byte
	ausfNC              nfdeployutil.NamingContext
)

func expectAusfCapacityProfile(mpsi *mps.MockPackageServiceInterface) {
	mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
		{
			ID:         1,
			ApiVersion: hydrationutil.IpAPIVersion,
			Kind:       "AusfCapacityProfile",
		},
	}), gomock.Eq(ausfNC)).Return(map[int][]string{
		1: {string(ausfcp)},
	}, nil).Times(1)
}

var _ = Describe("Ausfdeploy", func() {
	var (
		mockCtrl *gomock.Controller
		mpsi     *mps.MockPackageServiceInterface
		adi      *nftypehydration.AusfDeployImpl
	)
	ctx := context.Background()

	ausfcp, _ = os.ReadFile("../testhelper/ausfcapacityprofile.yaml")
	ausfDeploy1, _ = os.ReadFile("../testhelper/ausfdeploy1.yaml")
	ausfNC, _ = nfdeployutil.NewNamingContext(ausfClusterName, ausfNfDeployName)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mpsi = mps.NewMockPackageServiceInterface(mockCtrl)
		adi = &nftypehydration.AusfDeployImpl{
			PS:  mpsi,
			Log: ctrl.Log.WithName("ausfdeploy"),
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Testing GenerateNfTypeDeploy for single ausf site", func() {
		site := deployv1alpha1.Site{
			Id:          "ausf1",
			ClusterName: ausfClusterName,
			NFType:      "ausf",
			NFTypeName:  "ausfsmall",
			NFVendor:    "casa",
			NFVersion:   "1.0",
		}
		Context("testing ausfdeploy", func() {
			BeforeEach(func() {
				expectAusfCapacityProfile(mpsi)
			})
			It("should process a single ausf and return ausfdeploy", func() {
				format.MaxLength = 0
				resp, err := adi.GenerateNfTypeDeploy(ctx, site, ausfNfDeployName)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).To(Equal(ausfDeploy1))
			})
		})
		Context("expecting no value from packageservice for GetNFProfiles", func() {
			BeforeEach(func() {
				mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
					{
						ID:         1,
						ApiVersion: hydrationutil.IpAPIVersion,
						Kind:       "AusfCapacityProfile",
					},
				}), gomock.Eq(udmNC)).Return(map[int][]string{
					1: {},
				}, nil).Times(1)
			})
			It("should process a single ausf and return error", func() {
				resp, err := adi.GenerateNfTypeDeploy(ctx, site, ausfNfDeployName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("error getting AusfCapacityProfile"))
				Expect(resp).To(BeNil())
			})
		})
		Context("expecting error from packageservice for GetNFProfiles", func() {
			expectedErr := errors.New("error from packageservice")
			It("should return an error while getting ausfCapacityProfile from packageservice", func() {
				mpsi.EXPECT().GetNFProfiles(gomock.Any(), gomock.Eq([]ps.GetResourceRequest{
					{
						ID:         1,
						ApiVersion: hydrationutil.IpAPIVersion,
						Kind:       "AusfCapacityProfile",
					},
				}), gomock.Eq(ausfNC)).Return(nil, expectedErr).Times(1)

				resp, err := adi.GenerateNfTypeDeploy(ctx, site, ausfNfDeployName)
				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())
				Expect(err.Error()).To(HaveSuffix(expectedErr.Error()))
			})
		})
	})
})
