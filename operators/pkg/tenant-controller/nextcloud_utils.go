package tenant_controller

import (
	"encoding/json"
	"errors"
	"fmt"

	resty "github.com/go-resty/resty/v2"
	"k8s.io/klog/v2"
)

// NcHandler defines the method need to interact with nextcloud
type NcHandler interface {
	GetUser(ncUsername string) (found bool, displayname *string, err error)
	CreateUser(ncUsername, ncPsw, displayname string) error
	UpdateUserData(username, param, value string) error
	DeleteUser(username string) error
}

// NcActor holds the info and methods to interact with nextcloud
type NcActor struct {
	Client   *resty.Client
	TnOpUser string
	TnOpPsw  string
	BaseURL  string
}

// reference https://docs.nextcloud.com/server/15/admin_manual/configuration_user/instruction_set_for_users.html#add-a-new-user

var ncHeaders = map[string]string{"OCS-APIRequest": "true"}

// GetUser gets the user with the corresponding username in nextcloud. It returns info about the existence of the user, the displayname of the user and if there are any errors
func (ncA *NcActor) GetUser(ncUsername string) (found bool, displayname *string, err error) {
	userEndpoint := ncA.buildOCSEndpoint(fmt.Sprintf("/users/%s", ncUsername))
	userRes, err := ncA.Client.R().SetBasicAuth(ncA.TnOpUser, ncA.TnOpPsw).SetHeaders(ncHeaders).Get(userEndpoint)

	if err != nil {
		klog.Errorf("Error during GET request when getting user %s in nextcloud: %s", ncUsername, err)
		return false, nil, err
	}

	statusCode, message, err := parseOCSResponseMeta(userRes.Body())
	switch {
	case err != nil:
		klog.Errorf("Error when parsing meta of nextcloud response of GET request to get user %s: %s", ncUsername, err)
		return false, nil, err
	case *statusCode == 100:
		klog.Infof("Nextcloud user %s already existed", ncUsername)
		displayname, err := parseOCSResponseData(userRes.Body())
		if err != nil {
			klog.Errorf("Error when parsing data of nextcloud response of GET request to get user %s: %s", ncUsername, err)
			return false, nil, err
		}
		return true, displayname, nil
	case *statusCode == 404:
		klog.Infof("Nextcloud user %s did not exist", ncUsername)
		return false, nil, nil
	default:
		klog.Errorf("Error when getting user info of nextcloud user %s: statusCode: %d message: %s", ncUsername, *statusCode, *message)
		return false, nil, errors.New(*message)
	}
}

// CreateUser creates a new user with the passed username, psw and displayname
func (ncA *NcActor) CreateUser(ncUsername, ncPsw, displayname string) error {
	usersURL := ncA.buildOCSEndpoint("/users")
	userData := map[string]string{"userid": ncUsername, "password": ncPsw, "displayname": displayname}

	userRes, err := ncA.Client.R().SetBasicAuth(ncA.TnOpUser, ncA.TnOpPsw).SetHeaders(ncHeaders).SetFormData(userData).Post(usersURL)
	if err != nil {
		klog.Errorf("Error during POST on /users when creating user %s in nextcloud: %s", ncUsername, err)
		return err
	}

	statusCode, message, err := parseOCSResponseMeta(userRes.Body())
	if err != nil {
		klog.Errorf("Error when parsing nextcloud response when creating user %s: %s", ncUsername, err)
		return err
	}
	if *statusCode == 100 {
		klog.Infof("Nextcloud user for %s created", ncUsername)
	} else {
		klog.Errorf("Error when creating user in nextcloud: statusCode %d, message: %s", *statusCode, *message)
		return errors.New(*message)
	}
	return nil
}

// UpdateUserData updates the param of the user with the username with the new value
func (ncA *NcActor) UpdateUserData(username, param, value string) error {
	userURL := ncA.buildOCSEndpoint(fmt.Sprintf("/users/%s", username))
	data := map[string]string{"key": param, "value": value}
	res, err := ncA.Client.R().SetBasicAuth(ncA.TnOpUser, ncA.TnOpPsw).SetHeaders(ncHeaders).SetFormData(data).Put(userURL)
	if err != nil {
		klog.Errorf("Error when sending request for %s update in nextcloud for user %s: %s", param, username, err)
		return err
	}
	statusCode, message, err := parseOCSResponseMeta(res.Body())
	if err != nil {
		klog.Errorf("Error when parsing response of %s update in nextcloud of user %s: %s", param, username, err)
		return err
	}
	if *statusCode != 100 {
		klog.Errorf("Error when updating %s in nextcloud of user %s, statusCode: %d, message: %s", param, username, *statusCode, *message)
		return errors.New(*message)
	}
	return nil
}

// DeleteUser user deletes the user with the corresponding username
func (ncA *NcActor) DeleteUser(username string) error {
	userURL := ncA.buildOCSEndpoint(fmt.Sprintf("/users/%s", username))
	res, err := ncA.Client.R().SetBasicAuth(ncA.TnOpUser, ncA.TnOpPsw).SetHeaders(ncHeaders).Delete(userURL)
	if err != nil {
		klog.Errorf("Error when sending DELETE request for deletion in nextcloud for user %s: %s", username, err)
		return err
	}
	statusCode, message, err := parseOCSResponseMeta(res.Body())
	if err != nil {
		klog.Errorf("Error when parsing response of delete in nextcloud of user %s: %s", username, err)
		return err
	}
	if *statusCode != 100 {
		// since DELETE request fails if user is already deleted (WTF?!?!?) need to check if user still exists
		if found, _, err := ncA.GetUser(username); err != nil {
			klog.Errorf("Error when checking if nextcloud user %s exists: %s", username, err)
			return err
		} else if found {
			klog.Errorf("Error when deleting nextcloud user %s, user still exists despite attempted deletion, statusCode: %d, message: %s", username, *statusCode, *message)
			return errors.New("nextcloud user still exists after deletion")
		}
		return nil
	}
	return nil
}

func parseOCSResponseMeta(respBody []byte) (parsedStatusCode *int, parsedMessage *string, err error) {
	ocsJSON, err := extractOCSResponse(respBody)
	if err != nil {
		klog.Error("Error when extracting OCS response")
		return nil, nil, err
	}
	metaJSON := ocsJSON["meta"].(map[string]interface{})
	statusCode := int(metaJSON["statuscode"].(float64))
	message := metaJSON["message"].(string)
	return &statusCode, &message, nil
}

func parseOCSResponseData(respBody []byte) (parsedDisplayname *string, err error) {
	ocsJSON, err := extractOCSResponse(respBody)
	if err != nil {
		klog.Error("Error when extracting OCS response")
		return nil, err
	}
	metaJSON := ocsJSON["data"].(map[string]interface{})
	displayname := metaJSON["displayname"].(string)
	return &displayname, nil
}

func extractOCSResponse(respBody []byte) (map[string]interface{}, error) {
	respJSON := make(map[string]interface{})
	err := json.Unmarshal(respBody, &respJSON)
	if err != nil {
		klog.Error("Error when un-marshaling OCS response")
		return nil, err
	}
	ocsJSON := respJSON["ocs"].(map[string]interface{})
	return ocsJSON, nil
}

func (ncA *NcActor) buildOCSEndpoint(path string) string {
	return fmt.Sprintf("%s/ocs/v1.php/cloud%s?format=json", ncA.BaseURL, path)
}
