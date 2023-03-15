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

package controllers

import (
	"context"
	"errors"
	"fmt"
	"time"

	types4 "github.com/nephio-project/common-lib/ausf"
	types2 "github.com/nephio-project/common-lib/nfdeploy"
	types3 "github.com/nephio-project/common-lib/udm"
	"github.com/nephio-project/edge-watcher/preprocessor"
	"github.com/nephio-project/nf-deploy-controller/api/v1alpha1"
	"github.com/nephio-project/nf-deploy-controller/tests/utils"
	"github.com/nephio-project/nf-deploy-controller/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	crNfDeployPath         = "../config/samples/nfdeploy_v1alpha1_nfdeploy.yaml"
	crCompleteNfDeployPath = "../config/samples/nfdeploy_with_all_nfs.yaml"
)

func generateUPFEdgeEvent(
	stalledStatus corev1.ConditionStatus, availableStatus corev1.ConditionStatus,
	readyStatus corev1.ConditionStatus, peeringStatus corev1.ConditionStatus,
	reconcilingStatus corev1.ConditionStatus, name string,
) preprocessor.Event {
	upfDeploy := types2.UpfDeploy{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{util.NFSiteIDLabel: name},
		},
		Status: types2.UpfDeployStatus{
			Conditions: []types2.NFCondition{
				{Type: types2.Stalled, Status: stalledStatus},
				{Type: types2.Reconciling, Status: reconcilingStatus},
				{Type: types2.Available, Status: availableStatus},
				{Type: types2.Peering, Status: peeringStatus},
				{Type: types2.Ready, Status: readyStatus},
			},
		},
	}
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&upfDeploy)
	Expect(err).To(
		BeNil(),
		"unable to convert UpfDeploy type to unstructured.Unstructured",
	)

	return preprocessor.Event{
		Key: preprocessor.RequestKey{Namespace: "upf", Kind: "UPFDeploy"},
		Object: &unstructured.Unstructured{
			Object: data,
		},
	}
}

func generateSMFEdgeEvent(
	stalledStatus corev1.ConditionStatus, availableStatus corev1.ConditionStatus,
	readyStatus corev1.ConditionStatus, peeringStatus corev1.ConditionStatus,
	reconcilingStatus corev1.ConditionStatus, name string,
) preprocessor.Event {
	smfDeploy := types2.SmfDeploy{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{util.NFSiteIDLabel: name},
		},
		Status: types2.SmfDeployStatus{
			Conditions: []types2.NFCondition{
				{Type: types2.Stalled, Status: stalledStatus},
				{Type: types2.Reconciling, Status: reconcilingStatus},
				{Type: types2.Available, Status: availableStatus},
				{Type: types2.Peering, Status: peeringStatus},
				{Type: types2.Ready, Status: readyStatus},
			},
		},
	}

	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&smfDeploy)
	Expect(err).To(
		BeNil(),
		"unable to convert SmfDeploy type to unstructured.Unstructured",
	)

	return preprocessor.Event{
		Key: preprocessor.RequestKey{Namespace: "smf", Kind: "SMFDeploy"},
		Object: &unstructured.Unstructured{
			Object: data,
		},
	}
}

func generateUDMEdgeEvent(
	stalledStatus corev1.ConditionStatus, availableStatus corev1.ConditionStatus,
	readyStatus corev1.ConditionStatus, peeringStatus corev1.ConditionStatus,
	reconcilingStatus corev1.ConditionStatus,
) preprocessor.Event {
	udmDeploy := types3.UdmDeploy{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "udm-dummy",
			Labels: map[string]string{util.NFSiteIDLabel: "udm-dummy"},
		},
		Status: types3.UdmDeployStatus{
			Conditions: []types2.NFCondition{
				{Type: types2.Stalled, Status: stalledStatus},
				{Type: types2.Reconciling, Status: reconcilingStatus},
				{Type: types2.Available, Status: availableStatus},
				{Type: types2.Peering, Status: peeringStatus},
				{Type: types2.Ready, Status: readyStatus},
			},
		},
	}

	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&udmDeploy)
	Expect(err).To(
		BeNil(),
		"unable to convert UdmDeploy type to unstructured.Unstructured",
	)

	return preprocessor.Event{
		Key: preprocessor.RequestKey{Namespace: "udm", Kind: "UDMDeploy"},
		Object: &unstructured.Unstructured{
			Object: data,
		},
	}
}

func generateAUSFEdgeEvent(
	stalledStatus corev1.ConditionStatus, availableStatus corev1.ConditionStatus,
	readyStatus corev1.ConditionStatus, peeringStatus corev1.ConditionStatus,
	reconcilingStatus corev1.ConditionStatus,
) preprocessor.Event {
	ausfDeploy := types4.AusfDeploy{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "ausf-dummy",
			Labels: map[string]string{util.NFSiteIDLabel: "ausf-dummy"},
		},
		Status: types4.AusfDeployStatus{
			Conditions: []types2.NFCondition{
				{Type: types2.Stalled, Status: stalledStatus},
				{Type: types2.Reconciling, Status: reconcilingStatus},
				{Type: types2.Available, Status: availableStatus},
				{Type: types2.Peering, Status: peeringStatus},
				{Type: types2.Ready, Status: readyStatus},
			},
		},
	}

	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&ausfDeploy)
	Expect(err).To(
		BeNil(),
		"unable to convert AusfDeploy type to unstructured.Unstructured",
	)
	return preprocessor.Event{
		Key: preprocessor.RequestKey{Namespace: "ausf", Kind: "AUSFDeploy"},
		Object: &unstructured.Unstructured{
			Object: data,
		},
	}
}

func getNfDeployCr(path string) (*v1alpha1.NfDeploy, error) {
	u, err := utils.ParseYaml(
		path, schema.GroupVersionKind{
			Group:   "nfdeploy.nephio.org",
			Version: "v1alpha1",
			Kind:    "NfDeploy",
		},
	)
	if err != nil {
		return nil, err
	}
	nfDeploy := &v1alpha1.NfDeploy{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u, nfDeploy)
	if err != nil {
		return nil, err
	}
	nfDeploy.Namespace = "default"
	return nfDeploy, nil
}

func executeAndTestEdgeEventSequence(
	edgeEvents []preprocessor.Event,
	finalExpectedStatus map[v1alpha1.NFDeployConditionType]corev1.ConditionStatus,
	nfDeployName string,
) {
	nfDeploy, err := getNfDeployCr(crNfDeployPath)
	Expect(err).NotTo(HaveOccurred())
	nfDeploy.Name = nfDeployName
	Expect(k8sClient.Create(context.TODO(), nfDeploy)).Should(Succeed())
	Eventually(fakeDeploymentManager.SignalChan).Should(Receive(nil))
	req := <-fakeDeploymentManager.SubscriptionReqChan
	req.Error <- nil
	for _, edgeEvent := range edgeEvents {
		req.Channel <- edgeEvent
	}
	var newNfDeploy v1alpha1.NfDeploy
	Eventually(
		func() map[v1alpha1.NFDeployConditionType]corev1.ConditionStatus {
			// fetching latest nfDeploy
			err := k8sClient.Get(
				ctx, types.NamespacedName{
					Namespace: nfDeploy.Namespace,
					Name:      nfDeploy.Name,
				}, &newNfDeploy,
			)
			if err != nil {
				return nil
			}
			newMap := make(map[v1alpha1.NFDeployConditionType]corev1.ConditionStatus)
			for _, c := range newNfDeploy.Status.Conditions {
				newMap[c.Type] = c.Status
			}
			return newMap
		},
	).Should(Equal(finalExpectedStatus))

}

var _ = Describe(
	"NfDeploy Controller", func() {

		Context(
			"When NfDeploy is created and updated", func() {
				It(
					"Should report to Deployment Manager", func() {
						nfDeploy, err := getNfDeployCr(crNfDeployPath)
						Expect(err).NotTo(HaveOccurred())
						nfDeploy.Name = "deployment-manager-report-test"
						Expect(k8sClient.Create(context.TODO(), nfDeploy)).Should(Succeed())
						Eventually(fakeDeploymentManager.SignalChan).Should(Receive(nil))
						req := <-fakeDeploymentManager.SubscriptionReqChan
						req.Error <- nil
						err = retry.RetryOnConflict(
							retry.DefaultRetry, func() error {
								// fetching latest nfDeploy
								var newNfDeploy v1alpha1.NfDeploy
								if err := k8sClient.Get(
									context.TODO(), types.NamespacedName{
										Namespace: nfDeploy.Namespace,
										Name:      nfDeploy.Name,
									}, &newNfDeploy,
								); err != nil {
									return err
								}
								newNfDeploy.Spec.Plmn.MCC = newNfDeploy.Spec.Plmn.MCC + 1
								err := k8sClient.Update(context.TODO(), &newNfDeploy)
								return err
							},
						)
						Expect(err).To(BeNil())
						Eventually(fakeDeploymentManager.SignalChan).Should(Receive(nil))
					},
				)
			},
		)
		Context(
			"When nfdeploy validation fails", func() {
				It(
					"Should never call deployment manager", func() {
						nfDeploy, err := getNfDeployCr(crNfDeployPath)
						Expect(err).NotTo(HaveOccurred())
						newSites := []v1alpha1.Site{nfDeploy.Spec.Sites[0]}
						nfDeploy.Spec.Sites = newSites
						nfDeploy.Name = "deployment-nfdeploy-validation"
						Expect(k8sClient.Create(context.TODO(), nfDeploy)).Should(Succeed())
						Consistently(fakeDeploymentManager.SignalChan).Should(Not(Receive(nil)))
					},
				)
			},
		)

		Context(
			"NfDeploy is created", func() {
				Context(
					"Hydration is successful", func() {
						It(
							"Should update the status", func() {
								nfDeploy, err := getNfDeployCr(crNfDeployPath)
								Expect(err).NotTo(HaveOccurred())
								nfDeploy.Name = "hydration-successful"

								Expect(
									k8sClient.Create(
										context.TODO(), nfDeploy,
									),
								).Should(Succeed())
								req := <-fakeDeploymentManager.SubscriptionReqChan
								req.Error <- nil
								var cond v1alpha1.NFDeployCondition
								var newNfDeploy v1alpha1.NfDeploy
								Eventually(
									func() (string, error) {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return "", err
										}
										conditionTypes := []v1alpha1.NFDeployConditionType{}
										for _, c := range newNfDeploy.Status.Conditions {
											conditionTypes = append(conditionTypes, c.Type)
											if c.Type == v1alpha1.DeploymentReconciling {
												cond = c
												break
											}
										}
										return cond.Reason, nil
									},
								).Should(Equal("AwaitingApproval"))
								Expect(cond.Status).To(Equal(corev1.ConditionTrue))
								Expect(cond.Message).To(ContainSubstring("resourceName"))
								Expect(cond.Message).To(ContainSubstring("operator-resourceName"))
								Expect(newNfDeploy.Status.ObservedGeneration).To(Equal(int32(nfDeploy.Generation)))
							},
						)
					},
				)

				Context(
					"Hydration has failed", func() {
						It(
							"Should update the status when hydrate failed", func() {
								nfDeploy, err := getNfDeployCr(crNfDeployPath)
								Expect(err).NotTo(HaveOccurred())
								nfDeploy.Name = "hydration-failed"

								expectedErr := errors.New("error from porch")
								Expect(
									k8sClient.Create(
										context.TODO(), nfDeploy,
									),
								).Should(Succeed())
								var condReconciling, condStalled v1alpha1.NFDeployCondition
								var newNfDeploy v1alpha1.NfDeploy
								Eventually(
									func() (corev1.ConditionStatus, error) {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return corev1.ConditionUnknown, err
										}
										conditionTypes := []v1alpha1.NFDeployConditionType{}
										for _, c := range newNfDeploy.Status.Conditions {
											conditionTypes = append(conditionTypes, c.Type)
											if c.Type == v1alpha1.DeploymentReconciling {
												condReconciling = c
											} else if c.Type == v1alpha1.DeploymentStalled {
												condStalled = c
											}
										}
										return condStalled.Status, nil
									},
								).Should(Equal(corev1.ConditionTrue))

								// For DeploymentReconciling
								Expect(condReconciling.Status).To(Equal(corev1.ConditionFalse))
								Expect(condReconciling.Reason).To(Equal("Stalled"))
								Expect(condReconciling.Message).To(ContainSubstring(expectedErr.Error()))

								// For DeploymentStalled
								Expect(condStalled.Reason).To(Equal("HydrationFailure"))
								Expect(condStalled.Message).To(ContainSubstring(expectedErr.Error()))
								k8sClient.Delete(ctx, &newNfDeploy)
							},
						)

						It(
							"Should update the status when create actuation package failed",
							func() {
								nfDeploy, err := getNfDeployCr(crNfDeployPath)
								Expect(err).NotTo(HaveOccurred())
								nfDeploy.Name = "actuation-failed"

								expectedErr := errors.New("error from porch")
								Expect(
									k8sClient.Create(
										context.TODO(), nfDeploy,
									),
								).Should(Succeed())
								var condReconciling, condStalled v1alpha1.NFDeployCondition
								var newNfDeploy v1alpha1.NfDeploy
								Eventually(
									func() (corev1.ConditionStatus, error) {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return corev1.ConditionUnknown, err
										}
										conditionTypes := []v1alpha1.NFDeployConditionType{}
										for _, c := range newNfDeploy.Status.Conditions {
											conditionTypes = append(conditionTypes, c.Type)
											if c.Type == v1alpha1.DeploymentReconciling {
												condReconciling = c
											} else if c.Type == v1alpha1.DeploymentStalled {
												condStalled = c
											}
										}
										return condStalled.Status, nil
									},
								).Should(Equal(corev1.ConditionTrue))

								// For DeploymentReconciling
								Expect(condReconciling.Status).To(Equal(corev1.ConditionFalse))
								Expect(condReconciling.Reason).To(Equal("Stalled"))
								Expect(condReconciling.Message).To(ContainSubstring(expectedErr.Error()))

								// For DeploymentStalled
								Expect(condStalled.Reason).To(Equal("HydrationFailure"))
								Expect(condStalled.Message).To(ContainSubstring(expectedErr.Error()))
								k8sClient.Delete(ctx, &newNfDeploy)
							},
						)
					},
				)
			},
		)

		Context(
			"NfDeploy is deleted", func() {
				It(
					"Should delete the nfDeploy resource if there is no error from packageservice",
					func() {
						nfDeploy, err := getNfDeployCr(crNfDeployPath)
						Expect(err).NotTo(HaveOccurred())
						nfDeploy.Name = "nfdeploy-deletion"

						Expect(k8sClient.Create(context.TODO(), nfDeploy)).Should(Succeed())
						var cond v1alpha1.NFDeployCondition
						var newNfDeploy v1alpha1.NfDeploy
						req := <-fakeDeploymentManager.SubscriptionReqChan
						req.Error <- nil
						Eventually(
							func() (string, error) {
								// fetching latest nfDeploy
								if err := k8sClient.Get(
									ctx, types.NamespacedName{
										Namespace: nfDeploy.Namespace,
										Name:      nfDeploy.Name,
									}, &newNfDeploy,
								); err != nil {
									return "", err
								}
								conditionTypes := []v1alpha1.NFDeployConditionType{}
								for _, c := range newNfDeploy.Status.Conditions {
									conditionTypes = append(conditionTypes, c.Type)
									if c.Type == v1alpha1.DeploymentReconciling {
										cond = c
										break
									}
								}
								return cond.Reason, nil
							},
						).Should(Equal("AwaitingApproval"))

						Expect(
							k8sClient.Get(
								ctx, types.NamespacedName{
									Namespace: nfDeploy.Namespace,
									Name:      nfDeploy.Name,
								}, &newNfDeploy,
							),
						).Should(Succeed())
						Expect(k8sClient.Delete(ctx, &newNfDeploy)).Should(Succeed())
						err = k8sClient.Get(
							ctx, types.NamespacedName{
								Namespace: nfDeploy.Namespace,
								Name:      nfDeploy.Name,
							}, &newNfDeploy,
						)
						Expect(err).NotTo(HaveOccurred())
						Expect(newNfDeploy.ObjectMeta.DeletionTimestamp.IsZero()).To(BeFalse())
						Eventually(
							func() error {
								return k8sClient.Get(
									ctx, types.NamespacedName{
										Namespace: nfDeploy.Namespace,
										Name:      nfDeploy.Name,
									}, &newNfDeploy,
								)
							},
						).ShouldNot(Succeed())
					},
				)

				It(
					"Should not delete the nfDeploy resource if there is an error from packageservice",
					func() {
						nfDeploy, err := getNfDeployCr(crNfDeployPath)
						Expect(err).NotTo(HaveOccurred())
						nfDeploy.Name = "nfdeploy-deletion-error"

						Expect(k8sClient.Create(context.TODO(), nfDeploy)).Should(Succeed())
						var cond v1alpha1.NFDeployCondition
						var newNfDeploy v1alpha1.NfDeploy
						req := <-fakeDeploymentManager.SubscriptionReqChan
						req.Error <- nil
						Eventually(
							func() (string, error) {
								// fetching latest nfDeploy
								if err := k8sClient.Get(
									ctx, types.NamespacedName{
										Namespace: nfDeploy.Namespace,
										Name:      nfDeploy.Name,
									}, &newNfDeploy,
								); err != nil {
									return "", err
								}
								conditionTypes := []v1alpha1.NFDeployConditionType{}
								for _, c := range newNfDeploy.Status.Conditions {
									conditionTypes = append(conditionTypes, c.Type)
									if c.Type == v1alpha1.DeploymentReconciling {
										cond = c
										break
									}
								}
								return cond.Reason, nil
							},
						).Should(Equal("AwaitingApproval"))

						Expect(
							k8sClient.Get(
								ctx, types.NamespacedName{
									Namespace: nfDeploy.Namespace,
									Name:      nfDeploy.Name,
								}, &newNfDeploy,
							),
						).Should(Succeed())
						Expect(k8sClient.Delete(ctx, &newNfDeploy)).To(Succeed())
						Expect(
							k8sClient.Get(
								ctx, types.NamespacedName{
									Namespace: nfDeploy.Namespace,
									Name:      nfDeploy.Name,
								}, &newNfDeploy,
							),
						).To(Succeed())
						// checking if the nfDeploy resource still has the finalizer as there shuold have been error
						// deleting the porch packages so the controller did not proceed to the next step of removing finalizer
						Expect(
							controllerutil.ContainsFinalizer(
								&newNfDeploy, nfDeployFinalizerName,
							),
						).To(BeTrue())
					},
				)
			},
		)

		Context(
			"Deployment - NFDeployController Integration", func() {

				Context(
					"When edge returns error during connection establishing", func() {
						It(
							"Should set nfdeploy resource status", func() {
								nfDeploy, _ := getNfDeployCr(crNfDeployPath)
								nfDeploy.Name = "connection-failure-test"

								Expect(
									k8sClient.Create(
										context.TODO(), nfDeploy,
									),
								).Should(Succeed())
								Eventually(fakeDeploymentManager.SignalChan).Should(Receive(nil))
								req := <-fakeDeploymentManager.SubscriptionReqChan
								req.Error <- fmt.Errorf("test error from edgewatcher")
								var newNfDeploy v1alpha1.NfDeploy

								Eventually(
									func() string {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return ""
										}
										conditionTypes := []v1alpha1.NFDeployConditionType{}
										for _, c := range newNfDeploy.Status.Conditions {
											conditionTypes = append(conditionTypes, c.Type)
											if c.Type == v1alpha1.DeploymentReconciling {
												return c.Reason
											}
										}
										return ""
									},
								).Should(Equal("EdgeConnectionFailure"))
								for _, value := range newNfDeploy.Status.Conditions {
									Expect(value.Status).To(Equal(corev1.ConditionUnknown))
									Expect(value.Reason).To(Equal("EdgeConnectionFailure"))
								}
							},
						)
					},
				)

				Context(
					"When edge channel closes unexpectedly during connection establishing",
					func() {
						It(
							"Should set nfdeploy resource status", func() {

								nfdeploy2, _ := getNfDeployCr(crNfDeployPath)
								nfdeploy2.Name = "edge-channel-test"
								Expect(
									k8sClient.Create(
										context.TODO(), nfdeploy2,
									),
								).Should(Succeed())
								Eventually(fakeDeploymentManager.SignalChan).Should(Receive(nil))
								req := <-fakeDeploymentManager.SubscriptionReqChan
								close(req.Error)
								var newNfDeploy v1alpha1.NfDeploy
								Eventually(
									func() string {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfdeploy2.Namespace,
												Name:      nfdeploy2.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return ""
										}
										conditionTypes := []v1alpha1.NFDeployConditionType{}
										for _, c := range newNfDeploy.Status.Conditions {
											conditionTypes = append(conditionTypes, c.Type)
											if c.Type == v1alpha1.DeploymentReconciling {
												return c.Reason
											}
										}
										return ""
									},
								).Should(Equal("EdgeConnectionFailure"))
								for _, value := range newNfDeploy.Status.Conditions {
									Expect(value.Status).To(Equal(corev1.ConditionUnknown))
									Expect(value.Reason).To(Equal("EdgeConnectionFailure"))
								}
							},
						)
					},
				)
				Context(
					"When edge channel closes unexpectedly while listening to edge events",
					func() {
						It(
							"Should set nfdeploy resource status", func() {
								nfDeploy, _ := getNfDeployCr(crNfDeployPath)
								nfDeploy.Name = "edge-connection-closed-test"
								Expect(
									k8sClient.Create(
										context.TODO(), nfDeploy,
									),
								).Should(Succeed())
								Eventually(fakeDeploymentManager.SignalChan).Should(Receive(nil))

								req := <-fakeDeploymentManager.SubscriptionReqChan
								req.Error <- nil
								close(req.SubscriberInfo.Channel)
								var newNfDeploy v1alpha1.NfDeploy
								Eventually(
									func() string {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return ""
										}
										conditionTypes := []v1alpha1.NFDeployConditionType{}
										for _, c := range newNfDeploy.Status.Conditions {
											conditionTypes = append(conditionTypes, c.Type)
											if c.Type == v1alpha1.DeploymentReconciling {
												return c.Reason
											}
										}
										return ""
									},
								).Should(Equal("EdgeConnectionBroken"))
								for _, value := range newNfDeploy.Status.Conditions {
									Expect(value.Status).To(Equal(corev1.ConditionUnknown))
									Expect(value.Reason).To(Equal("EdgeConnectionBroken"))
								}
							},
						)
					},
				)

				Context(
					"When edge event with ambiguous condition set is provided", func() {
						It(
							"Should not update nfdeploy resource status", func() {
								nfDeploy, _ := getNfDeployCr(crNfDeployPath)
								nfDeploy.Name = "ambiguous-test"
								Expect(
									k8sClient.Create(
										context.TODO(), nfDeploy,
									),
								).Should(Succeed())
								Eventually(fakeDeploymentManager.SignalChan).Should(Receive(nil))

								req := <-fakeDeploymentManager.SubscriptionReqChan
								req.Error <- nil
								req.SubscriberInfo.Channel <- generateUPFEdgeEvent(
									corev1.ConditionUnknown, corev1.ConditionUnknown,
									corev1.ConditionUnknown, corev1.ConditionUnknown,
									corev1.ConditionUnknown, "upf-dummy",
								)
								var newNfDeploy v1alpha1.NfDeploy
								Consistently(
									func() string {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return ""
										}

										conditionTypes := []v1alpha1.NFDeployConditionType{}
										for _, c := range newNfDeploy.Status.Conditions {
											conditionTypes = append(conditionTypes, c.Type)
											if c.Type == v1alpha1.DeploymentReconciling {
												return c.Reason
											}
										}
										return ""
									},
								).Should(Equal("AwaitingApproval"))

							},
						)
					},
				)

				Context(
					"Testing reconciling status for valid edge events", func() {
						It(
							"Should update nfdeploy status", func() {
								// Two NFs in test resource
								nfDeploy, _ := getNfDeployCr(crNfDeployPath)
								nfDeploy.Name = "reconciling-test"
								Expect(
									k8sClient.Create(
										context.TODO(), nfDeploy,
									),
								).Should(Succeed())
								Eventually(fakeDeploymentManager.SignalChan).Should(Receive(nil))

								req := <-fakeDeploymentManager.SubscriptionReqChan
								req.Error <- nil

								// first NF reconciling
								req.SubscriberInfo.Channel <- generateUPFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionTrue, "upf-dummy",
								)
								var newNfDeploy v1alpha1.NfDeploy
								Eventually(
									func() string {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return ""
										}

										for _, c := range newNfDeploy.Status.Conditions {
											if c.Type == v1alpha1.DeploymentReconciling {
												return c.Reason
											}
										}
										return ""
									},
								).Should(Equal("SomeNFsReconciling"))

								// second NF reconciling
								req.SubscriberInfo.Channel <- generateSMFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionTrue, "smf-dummy",
								)
								Eventually(
									func() string {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return ""
										}

										for _, c := range newNfDeploy.Status.Conditions {
											if c.Type == v1alpha1.DeploymentReconciling {
												return c.Reason
											}
										}
										return ""
									},
								).Should(Equal("AllUnReconciledNFsReconciling"))

								// No NF reconciling
								req.SubscriberInfo.Channel <- generateUPFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse, "upf-dummy",
								)
								req.SubscriberInfo.Channel <- generateSMFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse, "smf-dummy",
								)
								Eventually(
									func() string {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return ""
										}

										for _, c := range newNfDeploy.Status.Conditions {
											if c.Type == v1alpha1.DeploymentReconciling {
												return c.Reason
											}
										}
										return ""
									},
								).Should(Equal("NoNFsReconciling"))

								// All NFs reconciled
								req.SubscriberInfo.Channel <- generateUPFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionTrue,
									corev1.ConditionTrue, corev1.ConditionFalse,
									corev1.ConditionFalse, "upf-dummy",
								)
								req.SubscriberInfo.Channel <- generateSMFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionTrue,
									corev1.ConditionTrue, corev1.ConditionFalse,
									corev1.ConditionFalse, "smf-dummy",
								)
								Eventually(
									func() string {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return ""
										}

										for _, c := range newNfDeploy.Status.Conditions {
											if c.Type == v1alpha1.DeploymentReconciling {
												return c.Reason
											}
										}
										return ""
									},
								).Should(Equal("AllNFsReconciled"))
							},
						)
					},
				)
				Context(
					"Testing peering status for valid edge events", func() {
						It(
							"Should update nfdeploy status", func() {
								// Two NFs in test resource
								nfDeploy, _ := getNfDeployCr(crNfDeployPath)
								nfDeploy.Name = "peering-test"
								Expect(
									k8sClient.Create(
										context.TODO(), nfDeploy,
									),
								).Should(Succeed())
								Eventually(fakeDeploymentManager.SignalChan).Should(Receive(nil))

								req := <-fakeDeploymentManager.SubscriptionReqChan
								req.Error <- nil

								// first NF peering
								req.SubscriberInfo.Channel <- generateUPFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionTrue,
									corev1.ConditionFalse, corev1.ConditionTrue,
									corev1.ConditionTrue, "upf-dummy",
								)
								var newNfDeploy v1alpha1.NfDeploy
								Eventually(
									func() string {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return ""
										}

										for _, c := range newNfDeploy.Status.Conditions {
											if c.Type == v1alpha1.DeploymentPeering {
												return c.Reason
											}
										}
										return ""
									},
								).Should(Equal("SomeNFsPeering"))

								// second NF peering
								req.SubscriberInfo.Channel <- generateSMFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionTrue,
									corev1.ConditionFalse, corev1.ConditionTrue,
									corev1.ConditionTrue, "smf-dummy",
								)
								Eventually(
									func() string {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return ""
										}

										for _, c := range newNfDeploy.Status.Conditions {
											if c.Type == v1alpha1.DeploymentPeering {
												return c.Reason
											}
										}
										return ""
									},
								).Should(Equal("AllUnPeeredNFsPeering"))

								// No NF peering
								req.SubscriberInfo.Channel <- generateUPFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse, "upf-dummy",
								)
								req.SubscriberInfo.Channel <- generateSMFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse, "smf-dummy",
								)
								Eventually(
									func() string {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return ""
										}

										for _, c := range newNfDeploy.Status.Conditions {
											if c.Type == v1alpha1.DeploymentPeering {
												return c.Reason
											}
										}
										return ""
									},
								).Should(Equal("NoNFsPeering"))

								// All NFs peered
								req.SubscriberInfo.Channel <- generateUPFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionTrue,
									corev1.ConditionTrue, corev1.ConditionFalse,
									corev1.ConditionFalse, "upf-dummy",
								)
								req.SubscriberInfo.Channel <- generateSMFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionTrue,
									corev1.ConditionTrue, corev1.ConditionFalse,
									corev1.ConditionFalse, "smf-dummy",
								)
								Eventually(
									func() string {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return ""
										}

										for _, c := range newNfDeploy.Status.Conditions {
											if c.Type == v1alpha1.DeploymentPeering {
												return c.Reason
											}
										}
										return ""
									},
								).Should(Equal("AllNFsPeered"))
							},
						)
					},
				)
				Context(
					"Testing ready status for valid edge events", func() {
						It(
							"Should update nfdeploy status", func() {
								// Two NFs in test resource
								nfDeploy, _ := getNfDeployCr(crNfDeployPath)
								nfDeploy.Name = "ready-test"
								Expect(
									k8sClient.Create(
										context.TODO(), nfDeploy,
									),
								).Should(Succeed())
								Eventually(fakeDeploymentManager.SignalChan).Should(Receive(nil))

								req := <-fakeDeploymentManager.SubscriptionReqChan
								req.Error <- nil

								// first NF ready
								req.SubscriberInfo.Channel <- generateUPFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionTrue,
									corev1.ConditionTrue, corev1.ConditionFalse,
									corev1.ConditionFalse, "upf-dummy",
								)
								var newNfDeploy v1alpha1.NfDeploy
								Eventually(
									func() string {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return ""
										}

										for _, c := range newNfDeploy.Status.Conditions {
											if c.Type == v1alpha1.DeploymentReady {
												return c.Reason
											}
										}
										return ""
									},
								).Should(Equal("SomeNFsReady"))

								// second NF ready
								req.SubscriberInfo.Channel <- generateSMFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionTrue,
									corev1.ConditionTrue, corev1.ConditionFalse,
									corev1.ConditionFalse, "smf-dummy",
								)
								Eventually(
									func() string {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return ""
										}

										for _, c := range newNfDeploy.Status.Conditions {
											if c.Type == v1alpha1.DeploymentReady {
												return c.Reason
											}
										}
										return ""
									},
								).Should(Equal("AllNFsReady"))

								// No NF ready
								req.SubscriberInfo.Channel <- generateUPFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse, "upf-dummy",
								)
								req.SubscriberInfo.Channel <- generateSMFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse, "smf-dummy",
								)
								Eventually(
									func() string {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return ""
										}

										for _, c := range newNfDeploy.Status.Conditions {
											if c.Type == v1alpha1.DeploymentReady {
												return c.Reason
											}
										}
										return ""
									},
								).Should(Equal("NoNFsReady"))

							},
						)
					},
				)
				Context(
					"Testing stalled status for valid edge events", func() {
						It(
							"Should update nfdeploy status", func() {
								// Two NFs in test resource
								nfDeploy, _ := getNfDeployCr(crNfDeployPath)
								nfDeploy.Name = "stalled-test"
								Expect(
									k8sClient.Create(
										context.TODO(), nfDeploy,
									),
								).Should(Succeed())
								Eventually(fakeDeploymentManager.SignalChan).Should(Receive(nil))

								req := <-fakeDeploymentManager.SubscriptionReqChan
								req.Error <- nil

								// first NF stalled
								req.SubscriberInfo.Channel <- generateUPFEdgeEvent(
									corev1.ConditionTrue, corev1.ConditionFalse,
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse, "upf-dummy",
								)
								req.SubscriberInfo.Channel <- generateSMFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionTrue,
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse, "smf-dummy",
								)
								var newNfDeploy v1alpha1.NfDeploy
								Eventually(
									func() string {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return ""
										}

										for _, c := range newNfDeploy.Status.Conditions {
											if c.Type == v1alpha1.DeploymentStalled {
												return c.Reason
											}
										}
										return ""
									},
								).Should(Equal("SomeNFsStalled"))

								// second NF stalled
								req.SubscriberInfo.Channel <- generateSMFEdgeEvent(
									corev1.ConditionTrue, corev1.ConditionFalse,
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse, "smf-dummy",
								)
								Eventually(
									func() string {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return ""
										}

										for _, c := range newNfDeploy.Status.Conditions {
											if c.Type == v1alpha1.DeploymentStalled {
												return c.Reason
											}
										}
										return ""
									},
								).Should(Equal("AllNFsStalled"))

								// No NF stalled
								req.SubscriberInfo.Channel <- generateUPFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionTrue,
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse, "upf-dummy",
								)
								req.SubscriberInfo.Channel <- generateSMFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionTrue,
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse, "smf-dummy",
								)
								Eventually(
									func() string {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return ""
										}

										for _, c := range newNfDeploy.Status.Conditions {
											if c.Type == v1alpha1.DeploymentStalled {
												return c.Reason
											}
										}
										return ""
									},
								).Should(Equal("NoNFsStalled"))

							},
						)
					},
				)
				Context(
					"Testing edge event sequence", func() {
						It(
							"Should update correct NFDeploy status", func() {
								var edgeEvents []preprocessor.Event
								edgeEvents = append(
									edgeEvents, generateUPFEdgeEvent(
										corev1.ConditionFalse, corev1.ConditionFalse,
										corev1.ConditionFalse, corev1.ConditionFalse,
										corev1.ConditionFalse, "upf-dummy",
									), generateUPFEdgeEvent(
										corev1.ConditionFalse, corev1.ConditionFalse,
										corev1.ConditionTrue, corev1.ConditionFalse,
										corev1.ConditionFalse, "upf-dummy",
									),
									generateUPFEdgeEvent(
										corev1.ConditionFalse, corev1.ConditionTrue,
										corev1.ConditionFalse, corev1.ConditionTrue,
										corev1.ConditionFalse, "upf-dummy",
									),
									generateUPFEdgeEvent(
										corev1.ConditionFalse, corev1.ConditionFalse,
										corev1.ConditionTrue, corev1.ConditionTrue,
										corev1.ConditionFalse, "upf-dummy",
									),
									generateUPFEdgeEvent(
										corev1.ConditionTrue, corev1.ConditionFalse,
										corev1.ConditionTrue, corev1.ConditionFalse,
										corev1.ConditionFalse, "upf-dummy",
									), generateUPFEdgeEvent(
										corev1.ConditionFalse, corev1.ConditionTrue,
										corev1.ConditionTrue, corev1.ConditionFalse,
										corev1.ConditionFalse, "upf-dummy",
									), generateSMFEdgeEvent(
										corev1.ConditionFalse, corev1.ConditionTrue,
										corev1.ConditionTrue, corev1.ConditionFalse,
										corev1.ConditionFalse, "smf-dummy",
									),
								)
								finalExpectedStatus := make(map[v1alpha1.NFDeployConditionType]corev1.ConditionStatus)
								finalExpectedStatus[v1alpha1.DeploymentStalled] = corev1.ConditionFalse
								finalExpectedStatus[v1alpha1.DeploymentPeering] = corev1.ConditionFalse
								finalExpectedStatus[v1alpha1.DeploymentReady] = corev1.ConditionTrue
								finalExpectedStatus[v1alpha1.DeploymentReconciling] = corev1.ConditionFalse
								executeAndTestEdgeEventSequence(
									edgeEvents, finalExpectedStatus, "edge-sequence-test-1",
								)
							},
						)
					},
				)
				Context(
					"Testing edge events with unknown status condition set", func() {
						It(
							"Should update correct NFDeploy status", func() {
								var edgeEvents []preprocessor.Event
								edgeEvents = append(
									edgeEvents, generateUPFEdgeEvent(
										corev1.ConditionFalse, corev1.ConditionFalse,
										corev1.ConditionFalse, corev1.ConditionFalse,
										corev1.ConditionFalse, "upf-dummy",
									), generateUPFEdgeEvent(
										corev1.ConditionUnknown, corev1.ConditionUnknown,
										corev1.ConditionTrue, corev1.ConditionUnknown,
										corev1.ConditionUnknown, "upf-dummy",
									), generateSMFEdgeEvent(
										corev1.ConditionTrue, corev1.ConditionUnknown,
										corev1.ConditionUnknown, corev1.ConditionUnknown,
										corev1.ConditionTrue, "smf-dummy",
									),
								)
								finalExpectedStatus := make(map[v1alpha1.NFDeployConditionType]corev1.ConditionStatus)
								finalExpectedStatus[v1alpha1.DeploymentStalled] = corev1.ConditionTrue
								finalExpectedStatus[v1alpha1.DeploymentPeering] = corev1.ConditionFalse
								finalExpectedStatus[v1alpha1.DeploymentReady] = corev1.ConditionFalse
								finalExpectedStatus[v1alpha1.DeploymentReconciling] = corev1.ConditionTrue
								executeAndTestEdgeEventSequence(
									edgeEvents, finalExpectedStatus, "edge-sequence-test-2",
								)
							},
						)
					},
				)
				Context(
					"Testing ausf and udm status for edge events", func() {
						It(
							"Should update nfdeploy status", func() {
								// Two NFs in test resource
								nfDeploy, _ := getNfDeployCr(crCompleteNfDeployPath)
								nfDeploy.Name = "ausf-udm-test"
								Expect(
									k8sClient.Create(
										context.TODO(), nfDeploy,
									),
								).Should(Succeed())
								Eventually(fakeDeploymentManager.SignalChan).Should(Receive(nil))

								req := <-fakeDeploymentManager.SubscriptionReqChan
								req.Error <- nil

								req.SubscriberInfo.Channel <- generateAUSFEdgeEvent(
									corev1.ConditionTrue, corev1.ConditionFalse,
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse,
								)
								req.SubscriberInfo.Channel <- generateUDMEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionTrue,
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse,
								)
								var newNfDeploy v1alpha1.NfDeploy
								Eventually(
									func() int32 {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return -1
										}

										return newNfDeploy.Status.StalledNFs
									},
								).Should(Equal(int32(1)))

								req.SubscriberInfo.Channel <- generateUDMEdgeEvent(
									corev1.ConditionTrue, corev1.ConditionFalse,
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse,
								)
								Eventually(
									func() int32 {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return -1
										}

										return newNfDeploy.Status.StalledNFs
									},
								).Should(Equal(int32(2)))

								// No NF stalled
								req.SubscriberInfo.Channel <- generateAUSFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionTrue,
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse,
								)
								req.SubscriberInfo.Channel <- generateUDMEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionTrue,
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse,
								)
								Eventually(
									func() int32 {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										if err != nil {
											return -1
										}

										return newNfDeploy.Status.StalledNFs
									},
								).Should(Equal(int32(0)))

							},
						)
					},
				)
			},
		)

		Context(
			"Deployment state cleanup on NFDeploy deletion", func() {
				When(
					"NFDeploy is deleted and created again with same name", func() {
						It(
							"Should clean deployment state", func() {
								nfDeploy, err := getNfDeployCr(crNfDeployPath)
								Expect(err).NotTo(HaveOccurred())
								nfDeploy.Name = "nfdeploy-deletion-2"
								Expect(
									k8sClient.Create(
										context.TODO(), nfDeploy,
									),
								).Should(Succeed())
								var newNfDeploy v1alpha1.NfDeploy
								req := <-fakeDeploymentManager.SubscriptionReqChan
								req.Error <- nil

								req.SubscriberInfo.Channel <- generateUPFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionTrue, "upf-dummy",
								)
								req.SubscriberInfo.Channel <- generateSMFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionUnknown,
									corev1.ConditionTrue, corev1.ConditionUnknown,
									corev1.ConditionUnknown, "smf-dummy",
								)
								Eventually(
									func(g Gomega) {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										g.Expect(err).To(BeNil())

										g.Expect((int)(newNfDeploy.Status.TargetedNFs)).To(Equal(len(nfDeploy.Spec.Sites)))
										var cond v1alpha1.NFDeployCondition
										for _, condition := range newNfDeploy.Status.Conditions {
											if condition.Type == v1alpha1.DeploymentReady {
												cond = condition
											}
										}
										g.Expect(cond.Reason).To(Equal("SomeNFsReady"))
									},
								).WithTimeout(time.Second * 12).
									WithPolling(time.Millisecond * 100).Should(Succeed())

								Expect(
									k8sClient.Get(
										ctx, types.NamespacedName{
											Namespace: nfDeploy.Namespace,
											Name:      nfDeploy.Name,
										}, &newNfDeploy,
									),
								).Should(Succeed())

								Expect(k8sClient.Delete(ctx, &newNfDeploy)).Should(Succeed())

								Eventually(
									func() error {
										return k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
									},
								).ShouldNot(Succeed())

								nfDeploy2, err := getNfDeployCr(crNfDeployPath)
								Expect(err).NotTo(HaveOccurred())
								nfDeploy2.Name = "nfdeploy-deletion-2"
								nfDeploy2.Spec.Sites[0].Id = "upf-dummy-2"
								nfDeploy2.Spec.Sites[1].Id = "smf-dummy-2"
								nfDeploy2.Spec.Sites[0].Connectivities[0].NeighborName = "smf-dummy-2"
								nfDeploy2.Spec.Sites[1].Connectivities[0].NeighborName = "upf-dummy-2"

								Expect(
									k8sClient.Create(
										context.TODO(), nfDeploy2,
									),
								).Should(Succeed())
								req = <-fakeDeploymentManager.SubscriptionReqChan
								req.Error <- nil
								req.SubscriberInfo.Channel <- generateUPFEdgeEvent(
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionFalse, corev1.ConditionFalse,
									corev1.ConditionTrue, "upf-dummy-2",
								)
								Eventually(
									func(g Gomega) {
										// fetching latest nfDeploy
										err := k8sClient.Get(
											ctx, types.NamespacedName{
												Namespace: nfDeploy.Namespace,
												Name:      nfDeploy.Name,
											}, &newNfDeploy,
										)
										g.Expect(err).To(BeNil())

										g.Expect((int)(newNfDeploy.Status.TargetedNFs)).To(Equal(len(nfDeploy2.Spec.Sites)))
										var cond v1alpha1.NFDeployCondition
										for _, condition := range newNfDeploy.Status.Conditions {
											if condition.Type == v1alpha1.DeploymentReady {
												cond = condition
											}
										}
										g.Expect(cond.Reason).To(Equal("NoNFsReady"))
									},
								).WithTimeout(time.Second * 12).
									WithPolling(time.Millisecond * 100).Should(Succeed())
							},
						)
					},
				)
			},
		)
	},
)
