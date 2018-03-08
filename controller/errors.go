/**
 * Copyright 2018 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controller

import (
	"fmt"
)

// TODO need to remove this error, since its moved to ubiquity it self
type NoMounterForVolumeError struct {
	mounter string
}

func (e *NoMounterForVolumeError) Error() string {
	return fmt.Sprintf("Mounter not found for backend: %s", e.mounter)
}

const MissingWwnMountRequestErrorStr = "volume related to scbe backend must have mountRequest.Opts[Wwn] not found (expect to have wwn for scbe backend volume type)"

const FailRemovePVorigDirErrorStr = "Failed removing existing volume directory"
type FailRemovePVorigDirError struct {
	err string
	dir error
}

func (e *FailRemovePVorigDirError) Error() string {
	return fmt.Sprintf(FailRemovePVorigDirErrorStr + ". dir=[%s], err=[%#v]", e.dir, e.err)
}

