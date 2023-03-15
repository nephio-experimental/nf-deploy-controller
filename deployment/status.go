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

package deployment

import (
	"context"
	"strings"
	"time"

	types "github.com/nephio-project/common-lib/nfdeploy"
	"github.com/nephio-project/nf-deploy-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

// isAmbiguousConditionSet: returns true when no condition's status is set to True
func (deployment *Deployment) isAmbiguousConditionSet(conditions map[types.NFConditionType]corev1.ConditionStatus) bool {
	isAnyConditionSet := false
	for _, status := range conditions {
		if status == corev1.ConditionTrue {
			isAnyConditionSet = true
		}
	}
	isStalled := false
	if conditions[types.Available] == corev1.ConditionFalse &&
		conditions[types.Reconciling] == corev1.ConditionFalse {
		isStalled = true
	}

	return !isAnyConditionSet && !isStalled
}

// validateNFConditionSet : checks if the given condition set from NF Status is
// valid or not. Updates NFStatus as stalled if the given condition is not valid
func (deployment *Deployment) validateNFConditionSet(
	conditions map[types.NFConditionType]corev1.ConditionStatus,

) (NFStatus, bool) {

	if conditions[types.Ready] == corev1.ConditionTrue &&
		conditions[types.Stalled] == corev1.ConditionTrue {
		message := "Inconsistent NFTypeDeploy status received. Ready and stalled " +
			"conditions cannot be true at the same time."
		return NFStatus{
			state:            types.Stalled,
			stateMessage:     message,
			activeConditions: map[types.NFConditionType]string{types.Stalled: message},
		}, true
	}
	if (conditions[types.Ready] == corev1.ConditionTrue ||
		conditions[types.Peering] == corev1.ConditionTrue) &&
		conditions[types.Available] == corev1.ConditionFalse {
		message := "Inconsistent NFTypeDeploy status received." +
			" Available condition cannot be false when ready or peering conditions are true."
		return NFStatus{
			state:            types.Stalled,
			stateMessage:     message,
			activeConditions: map[types.NFConditionType]string{types.Stalled: message},
		}, true

	}
	if conditions[types.Reconciling] == corev1.ConditionFalse && conditions[types.Peering] == corev1.ConditionTrue {
		message := "Inconsistent NFTypeDeploy status received. Reconciling " +
			"condition cannot be false when peering condition is true."
		return NFStatus{
			state:            types.Stalled,
			stateMessage:     message,
			activeConditions: map[types.NFConditionType]string{types.Stalled: message},
		}, true
	}
	if (conditions[types.Peering] == corev1.ConditionTrue ||
		conditions[types.Reconciling] == corev1.ConditionTrue) &&
		conditions[types.Ready] == corev1.ConditionTrue {
		message := "Inconsistent NFTypeDeploy status received." +
			" Ready condition cannot be true when either reconciling or peering " +
			"condition are true."
		return NFStatus{
			state:            types.Stalled,
			stateMessage:     message,
			activeConditions: map[types.NFConditionType]string{types.Stalled: message},
		}, true
	}

	return NFStatus{}, false
}

// computeReconcilingCondition : computes DeploymentReconciling NFDeployCondition
// from status of present NFs in deployment
func (deployment *Deployment) computeReconcilingCondition(
	readyNFs int, targetedNFs int,
) v1alpha1.NFDeployCondition {
	reconcilingCondition := v1alpha1.NFDeployCondition{}
	reconcilingCondition.Type = v1alpha1.DeploymentReconciling

	if targetedNFs == readyNFs {
		reconcilingCondition.Reason = "AllNFsReconciled"
		reconcilingCondition.Status = corev1.ConditionFalse
		reconcilingCondition.Message = "All NFs are in reconciled state."
		return reconcilingCondition
	}
	reconcilingNFs := 0
	message := "The NFs which are in Reconciling state are: "
	for _, node := range deployment.upfNodes {
		if _, isPresent := node.Status.activeConditions[types.Reconciling]; isPresent {
			message = message + node.Id + ", "
			reconcilingNFs++
		}
	}
	for _, node := range deployment.smfNodes {
		if _, isPresent := node.Status.activeConditions[types.Reconciling]; isPresent {
			message = message + node.Id + ", "
			reconcilingNFs++
		}
	}
	for _, node := range deployment.ausfNodes {
		if _, isPresent := node.Status.activeConditions[types.Reconciling]; isPresent {
			message = message + node.Id + ", "
			reconcilingNFs++
		}
	}
	for _, node := range deployment.udmNodes {
		if _, isPresent := node.Status.activeConditions[types.Reconciling]; isPresent {
			message = message + node.Id + ", "
			reconcilingNFs++
		}
	}
	if readyNFs+reconcilingNFs == targetedNFs {
		reconcilingCondition.Reason = "AllUnReconciledNFsReconciling"
		message = strings.TrimSuffix(message, ", ")
		message = message + "."
		reconcilingCondition.Message = message
		reconcilingCondition.Status = corev1.ConditionTrue
	} else if reconcilingNFs != 0 {
		reconcilingCondition.Reason = "SomeNFsReconciling"
		message = strings.TrimSuffix(message, ", ")
		message = message + "."
		reconcilingCondition.Message = message
		reconcilingCondition.Status = corev1.ConditionTrue

	} else {
		reconcilingCondition.Reason = "NoNFsReconciling"
		reconcilingCondition.Message = "No NFs are in reconciling state."
		reconcilingCondition.Status = corev1.ConditionFalse

	}
	return reconcilingCondition

}

// computePeeringCondition : computes DeploymentPeering NFDeployCondition
// from status of present NFs in deployment
func (deployment *Deployment) computePeeringCondition(
	readyNFs int, targetedNFs int,
) v1alpha1.NFDeployCondition {
	peeringCondition := v1alpha1.NFDeployCondition{}
	peeringCondition.Type = v1alpha1.DeploymentPeering

	if targetedNFs == readyNFs {
		peeringCondition.Reason = "AllNFsPeered"
		peeringCondition.Status = corev1.ConditionFalse
		peeringCondition.Message = "All NFs are in Peered state."
		return peeringCondition
	}
	peeringNFs := 0
	message := "The NFs which are in Peering state are: "
	for _, node := range deployment.upfNodes {
		if _, isPresent := node.Status.activeConditions[types.Peering]; isPresent {
			message = message + node.Id + ", "
			peeringNFs++
		}
	}
	for _, node := range deployment.smfNodes {
		if _, isPresent := node.Status.activeConditions[types.Peering]; isPresent {
			message = message + node.Id + ", "
			peeringNFs++
		}
	}
	for _, node := range deployment.ausfNodes {
		if _, isPresent := node.Status.activeConditions[types.Peering]; isPresent {
			message = message + node.Id + ", "
			peeringNFs++
		}
	}
	for _, node := range deployment.udmNodes {
		if _, isPresent := node.Status.activeConditions[types.Peering]; isPresent {
			message = message + node.Id + ", "
			peeringNFs++
		}
	}
	if readyNFs+peeringNFs == targetedNFs {
		peeringCondition.Reason = "AllUnPeeredNFsPeering"
		message = strings.TrimSuffix(message, ", ")
		message = message + "."
		peeringCondition.Message = message
		peeringCondition.Status = corev1.ConditionTrue

	} else if peeringNFs != 0 {
		peeringCondition.Reason = "SomeNFsPeering"
		message = strings.TrimSuffix(message, ", ")
		message = message + "."
		peeringCondition.Message = message
		peeringCondition.Status = corev1.ConditionTrue

	} else {
		peeringCondition.Reason = "NoNFsPeering"
		peeringCondition.Message = "No NFs are in peering state."
		peeringCondition.Status = corev1.ConditionFalse

	}
	return peeringCondition

}

// computeReadyCondition : computes DeploymentReady NFDeployCondition
// from status of present NFs in deployment
func (deployment *Deployment) computeReadyCondition(
	readyNFs int, targetedNFs int,
) v1alpha1.NFDeployCondition {
	readyCondition := v1alpha1.NFDeployCondition{}
	readyCondition.Type = v1alpha1.DeploymentReady
	if readyNFs == targetedNFs {
		readyCondition.Status = corev1.ConditionTrue
	} else {
		readyCondition.Status = corev1.ConditionFalse
	}
	if readyNFs == 0 {
		readyCondition.Reason = "NoNFsReady"
	} else if readyNFs == targetedNFs {
		readyCondition.Reason = "AllNFsReady"
	} else {
		readyCondition.Reason = "SomeNFsReady"
	}
	if readyNFs == targetedNFs {
		readyCondition.Message = "All NFs are in Ready state."
		return readyCondition
	}
	message := "The NFs which are not in Ready state are: "
	for _, node := range deployment.upfNodes {
		if _, isPresent := node.Status.activeConditions[types.Ready]; !isPresent {
			message = message + node.Id + ", "
		}
	}
	for _, node := range deployment.smfNodes {
		if _, isPresent := node.Status.activeConditions[types.Ready]; !isPresent {
			message = message + node.Id + ", "
		}
	}
	for _, node := range deployment.ausfNodes {
		if _, isPresent := node.Status.activeConditions[types.Ready]; !isPresent {
			message = message + node.Id + ", "
		}
	}
	for _, node := range deployment.udmNodes {
		if _, isPresent := node.Status.activeConditions[types.Ready]; !isPresent {
			message = message + node.Id + ", "
		}
	}
	message = strings.TrimSuffix(message, ", ")
	message = message + "."
	readyCondition.Message = message
	return readyCondition
}

// computeStalledCondition : computes DeploymentStalled NFDeployCondition
// from status of present NFs in deployment
func (deployment *Deployment) computeStalledCondition(
	stalledNFs int, targetedNFs int,
) v1alpha1.NFDeployCondition {
	stalledCondition := v1alpha1.NFDeployCondition{}
	stalledCondition.Type = v1alpha1.DeploymentStalled
	if stalledNFs != 0 {
		stalledCondition.Status = corev1.ConditionTrue
	} else {
		stalledCondition.Status = corev1.ConditionFalse
	}
	if stalledNFs == targetedNFs {
		stalledCondition.Reason = "AllNFsStalled"
	} else if stalledNFs == 0 {
		stalledCondition.Reason = "NoNFsStalled"
	} else {
		stalledCondition.Reason = "SomeNFsStalled"
	}
	if stalledNFs == 0 {
		stalledCondition.Message = "No NFs are in stalled state."
		return stalledCondition
	}
	message := ""
	for _, node := range deployment.upfNodes {
		for conditionType, conditionStatus := range node.Status.activeConditions {
			if conditionType == types.Stalled {
				message = message + node.Id + ": " + conditionStatus + ", "
			}
		}
	}
	for _, node := range deployment.smfNodes {
		for conditionType, conditionStatus := range node.Status.activeConditions {
			if conditionType == types.Stalled {
				message = message + node.Id + ": " + conditionStatus + ", "
			}
		}
	}
	for _, node := range deployment.ausfNodes {
		for conditionType, conditionStatus := range node.Status.activeConditions {
			if conditionType == types.Stalled {
				message = message + node.Id + ": " + conditionStatus + ", "
			}
		}
	}
	for _, node := range deployment.udmNodes {
		for conditionType, conditionStatus := range node.Status.activeConditions {
			if conditionType == types.Stalled {
				message = message + node.Id + ": " + conditionStatus + ", "
			}
		}
	}
	message = strings.TrimSuffix(message, ", ")
	message = message + "."
	stalledCondition.Message = message
	return stalledCondition
}

// calculateNFCount: Calculate count of available, ready, stalled and targeted
// NFs present in a deployment
func (deployment *Deployment) calculateNFCount() (int, int, int, int) {
	availableNFs := 0
	readyNFs := 0
	stalledNFs := 0
	targetedNFs := len(deployment.upfNodes) + len(deployment.smfNodes) +
		len(deployment.ausfNodes) + len(deployment.udmNodes)
	for _, node := range deployment.upfNodes {
		for conditionType := range node.Status.activeConditions {
			if conditionType == types.Ready {
				readyNFs++
			}
			if conditionType == types.Stalled {
				stalledNFs++
			}
			if conditionType == types.Available {
				availableNFs++
			}
		}
	}
	for _, node := range deployment.smfNodes {
		for conditionType := range node.Status.activeConditions {
			if conditionType == types.Ready {
				readyNFs++
			}
			if conditionType == types.Stalled {
				stalledNFs++
			}
			if conditionType == types.Available {
				availableNFs++
			}
		}
	}
	for _, node := range deployment.ausfNodes {
		for conditionType := range node.Status.activeConditions {
			if conditionType == types.Ready {
				readyNFs++
			}
			if conditionType == types.Stalled {
				stalledNFs++
			}
			if conditionType == types.Available {
				availableNFs++
			}
		}
	}
	for _, node := range deployment.udmNodes {
		for conditionType := range node.Status.activeConditions {
			if conditionType == types.Ready {
				readyNFs++
			}
			if conditionType == types.Stalled {
				stalledNFs++
			}
			if conditionType == types.Available {
				availableNFs++
			}
		}
	}

	return availableNFs, readyNFs, stalledNFs, targetedNFs
}

// updateCurrentNFStatus: computes and updates in memory status of an NFNode
// present in current deployment
func (deployment *Deployment) updateCurrentNFStatus(
	nfId string, conditions map[types.NFConditionType]corev1.ConditionStatus,
	conditionMessage map[types.NFConditionType]string,
) {
	currentStatus, isSet := deployment.validateNFConditionSet(conditions)

	if !isSet {
		activeConditions := make(map[types.NFConditionType]string)
		for conditionType, conditionStatus := range conditions {
			if conditionStatus == corev1.ConditionTrue {
				activeConditions[conditionType] = conditionMessage[conditionType]
			}
		}
		if conditions[types.Available] == corev1.ConditionFalse &&
			conditions[types.Reconciling] == corev1.ConditionFalse {
			message := "NF is neither available nor reconciling."
			currentStatus = NFStatus{
				state:            types.Stalled,
				stateMessage:     message,
				activeConditions: map[types.NFConditionType]string{types.Stalled: message},
			}
		} else if conditions[types.Stalled] == corev1.ConditionTrue {

			currentStatus = NFStatus{
				state: types.Stalled, stateMessage: conditionMessage[types.Stalled],
				activeConditions: activeConditions,
			}
		} else if conditions[types.Ready] == corev1.ConditionTrue {
			if _, isPresent := activeConditions[types.Available]; !isPresent {
				activeConditions[types.Available] = "NF is in ready state."
			}
			currentStatus = NFStatus{
				state: types.Ready, stateMessage: conditionMessage[types.Ready],
				activeConditions: activeConditions,
			}
		} else if conditions[types.Peering] == corev1.ConditionTrue {
			if _, isPresent := activeConditions[types.Reconciling]; !isPresent {
				activeConditions[types.Reconciling] = "NF is in peering state."
			}
			if _, isPresent := activeConditions[types.Available]; !isPresent {
				activeConditions[types.Available] = "NF is in peering state."
			}
			currentStatus = NFStatus{
				state: types.Peering, stateMessage: conditionMessage[types.Peering],
				activeConditions: activeConditions,
			}
		} else if conditions[types.Reconciling] == corev1.ConditionTrue {
			currentStatus = NFStatus{
				state:            types.Reconciling,
				stateMessage:     conditionMessage[types.Reconciling],
				activeConditions: activeConditions,
			}
		} else {
			currentStatus = NFStatus{
				state:            types.Available,
				stateMessage:     conditionMessage[types.Available],
				activeConditions: activeConditions,
			}
		}
	}
	if _, isPresent := deployment.upfNodes[nfId]; isPresent {
		nf := deployment.upfNodes[nfId]
		currentStatus.lastEventTimestamp = nf.Status.lastEventTimestamp
		nf.Status = currentStatus
		deployment.upfNodes[nfId] = nf
	}
	if _, isPresent := deployment.smfNodes[nfId]; isPresent {
		nf := deployment.smfNodes[nfId]
		currentStatus.lastEventTimestamp = nf.Status.lastEventTimestamp
		nf.Status = currentStatus
		deployment.smfNodes[nfId] = nf
	}
	if _, isPresent := deployment.ausfNodes[nfId]; isPresent {
		nf := deployment.ausfNodes[nfId]
		currentStatus.lastEventTimestamp = nf.Status.lastEventTimestamp
		nf.Status = currentStatus
		deployment.ausfNodes[nfId] = nf
	}
	if _, isPresent := deployment.udmNodes[nfId]; isPresent {
		nf := deployment.udmNodes[nfId]
		currentStatus.lastEventTimestamp = nf.Status.lastEventTimestamp
		nf.Status = currentStatus
		deployment.udmNodes[nfId] = nf
	}
}

// calculateNFConditionSet: Returns maps which store the Status and Message of
// all NFConditions present in NFStatus. Assumes Unknown Status of all
// absent conditions
func (deployment *Deployment) calculateNFConditionSet(nfConditions *[]types.NFCondition) (
	map[types.NFConditionType]corev1.ConditionStatus,
	map[types.NFConditionType]string,
) {
	var conditions = make(map[types.NFConditionType]corev1.ConditionStatus)
	var conditionMessage = make(map[types.NFConditionType]string)
	for _, condition := range *nfConditions {
		conditions[condition.Type] = condition.Status
		conditionMessage[condition.Type] = condition.Message
	}
	if _, isPresent := conditions[types.Reconciling]; !isPresent {
		conditions[types.Reconciling] = corev1.ConditionUnknown
	}
	if _, isPresent := conditions[types.Peering]; !isPresent {
		conditions[types.Peering] = corev1.ConditionUnknown
	}
	if _, isPresent := conditions[types.Ready]; !isPresent {
		conditions[types.Ready] = corev1.ConditionUnknown
	}
	if _, isPresent := conditions[types.Stalled]; !isPresent {
		conditions[types.Stalled] = corev1.ConditionUnknown
	}
	if _, isPresent := conditions[types.Available]; !isPresent {
		conditions[types.Available] = corev1.ConditionUnknown
	}
	return conditions, conditionMessage
}

// areNFDeployStatusConditionsEqual: Returns true if two conditions have same
// Status, Reason and Message
func (deployment *Deployment) areNFDeployStatusConditionsEqual(
	conditionA *v1alpha1.NFDeployCondition,
	conditionB v1alpha1.NFDeployCondition,
) bool {
	if conditionA.Reason == conditionB.Reason && conditionA.Status == conditionB.Status &&
		conditionA.Message == conditionB.Message {
		return true
	}
	return false
}

// computeSingleNFDeployStatusConditionChange: updates the LastUpdateTime and
// LastTransitionTime based on change in condition currently computed and the
// one received from api-server
func (deployment *Deployment) computeSingleNFDeployStatusConditionChange(
	conditions []v1alpha1.NFDeployCondition,
	currentCondition *v1alpha1.NFDeployCondition,
) {
	conditionIdx := -1
	for index, originalCondition := range conditions {
		if originalCondition.Type == currentCondition.Type {
			conditionIdx = index
		}
	}
	currentTime := metav1.NewTime(time.Now())
	if conditionIdx == -1 {
		currentCondition.LastTransitionTime = currentTime
		currentCondition.LastUpdateTime = currentTime
	} else if deployment.areNFDeployStatusConditionsEqual(
		currentCondition, conditions[conditionIdx],
	) {
		currentCondition = &conditions[conditionIdx]
	} else {
		if currentCondition.Status == conditions[conditionIdx].Status {
			currentCondition.LastTransitionTime = conditions[conditionIdx].LastTransitionTime
		} else {
			currentCondition.LastTransitionTime = currentTime
		}
		currentCondition.LastUpdateTime = currentTime
	}

}

// updateNFDeployStatus: computes NFDeployCondition change and updates status of
// nfdeploy resource which the deployment is tracking. Returns error if update
// fails after exhausting retries or receiving a non-retryable error
func (deployment *Deployment) updateNFDeployStatus(
	availableNFs int32, readyNFs int32, stalledNFs int32, targetedNFs int32,
	stalledCondition *v1alpha1.NFDeployCondition,
	readyCondition *v1alpha1.NFDeployCondition,
	peeringCondition *v1alpha1.NFDeployCondition,
	reconcilingCondition *v1alpha1.NFDeployCondition,
) error {
	err := retry.RetryOnConflict(
		retry.DefaultRetry, func() error {
			var nfDeploy v1alpha1.NfDeploy
			if err := deployment.statusReader.Get(
				context.TODO(), deployment.namespacedName, &nfDeploy,
			); err != nil {
				return err
			}
			deployment.computeSingleNFDeployStatusConditionChange(
				nfDeploy.Status.Conditions, stalledCondition,
			)
			deployment.computeSingleNFDeployStatusConditionChange(
				nfDeploy.Status.Conditions, readyCondition,
			)
			deployment.computeSingleNFDeployStatusConditionChange(
				nfDeploy.Status.Conditions, peeringCondition,
			)
			deployment.computeSingleNFDeployStatusConditionChange(
				nfDeploy.Status.Conditions, reconcilingCondition,
			)
			newConditions := []v1alpha1.NFDeployCondition{
				*stalledCondition, *readyCondition, *peeringCondition,
				*reconcilingCondition,
			}
			newNFDeployStatus := v1alpha1.NfDeployStatus{
				ObservedGeneration: nfDeploy.Status.ObservedGeneration,
				TargetedNFs:        targetedNFs, ReadyNFs: readyNFs,
				AvailableNFs: availableNFs, StalledNFs: stalledNFs,
				Conditions: newConditions,
			}
			nfDeploy.Status = newNFDeployStatus
			if err := deployment.statusWriter.Update(
				context.TODO(), &nfDeploy,
			); err != nil {
				return err
			}
			return nil
		},
	)

	return err
}
