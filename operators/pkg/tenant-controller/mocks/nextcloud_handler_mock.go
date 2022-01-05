// Copyright 2020-2022 Politecnico di Torino
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mocks

// NcHandlerMock mocks NcHandler for testing nextcloud features.
type NcHandlerMock struct {
}

// GetUser mocks GetUser by implementing only one case, the user already exists, with displayname and there are no errors.
func (mNcA *NcHandlerMock) GetUser(ncUsername string) (found bool, displayname *string, err error) {
	mDisplayname := "displaname"
	return true, &mDisplayname, nil
}

// CreateUser mocks CreateUser by implementing only one case, the creation is successful.
func (mNcA *NcHandlerMock) CreateUser(ncUsername, ncPsw, displayname string) error { return nil }

// UpdateUserData mocks UpdateUserData by implementing only one case, the creation is successful.
func (mNcA *NcHandlerMock) UpdateUserData(username, param, value string) error { return nil }

// DeleteUser mocks DeleteUser by implementing only one case, the creation is successful.
func (mNcA *NcHandlerMock) DeleteUser(username string) error { return nil }
