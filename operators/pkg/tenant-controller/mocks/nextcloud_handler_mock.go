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
