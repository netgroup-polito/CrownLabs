// Copyright 2020-2025 Politecnico di Torino
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package shvolctrl

const (
	// EvPVCSmaller -> the event key corresponding to an invalid resizing.
	EvPVCSmaller = "InvalidSize"
	// EvPVCSmallerMsg -> the event message corresponding to an invalid resizing.
	EvPVCSmallerMsg = "Size cannot be less than previous value"

	// EvPVNoCSI -> the event key corresponding to missing CSI params.
	EvPVNoCSI = "MissingCSI"
	// EvPVNoCSIMsg -> the event message corresponding to missing CSI params.
	EvPVNoCSIMsg = "PV misses CSI params"

	// EvPVCResQuotaExceeded -> the event key corresponding to exceeded PVC quota.
	EvPVCResQuotaExceeded = "ResourceQuotaExceeded"
	// EvPVCResQuotaExceededMsg -> the event message corresponding to exceeded PVC quota.
	EvPVCResQuotaExceededMsg = "PVC exceeded the resource quota"

	// EvDeletionBlocked -> the event key corresponding to blocked deletion.
	EvDeletionBlocked = "DeletionBlocked"
	// EvDeletionBlockedMsg -> the event message corresponding to blocked deletion.
	EvDeletionBlockedMsg = "Cannot delete shvol since it is mounted on %v"
)
