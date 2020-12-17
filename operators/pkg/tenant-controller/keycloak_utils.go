package tenant_controller

import (
	"context"
	"errors"
	"fmt"
	"strings"

	gocloak "github.com/Nerzal/gocloak/v7"
	"k8s.io/klog"
)

// KcActor contains the needed objects and infos to use keycloak functionalities
type KcActor struct {
	Client                gocloak.GoCloak
	Token                 *gocloak.JWT
	TargetRealm           string
	TargetClientID        string
	UserRequiredActions   []string
	EmailActionsLifeSpanS int
}

func (kcA *KcActor) createKcRoles(ctx context.Context, rolesToCreate []string) error {
	for _, newRoleName := range rolesToCreate {
		if err := kcA.createKcRole(ctx, newRoleName); err != nil {
			klog.Error("Could not create user role", newRoleName)
			return err
		}
	}
	return nil
}

func (kcA *KcActor) createKcRole(ctx context.Context, newRoleName string) error {
	// check if keycloak role already esists

	role, err := kcA.Client.GetClientRole(ctx, kcA.Token.AccessToken, kcA.TargetRealm, kcA.TargetClientID, newRoleName)
	if err != nil && strings.Contains(err.Error(), "Could not find role") {
		// role didn't exist
		// need to create new role
		klog.Infof("Role didn't exist %s", newRoleName)
		tr := true
		createdRoleName, err := kcA.Client.CreateClientRole(ctx, kcA.Token.AccessToken, kcA.TargetRealm, kcA.TargetClientID, gocloak.Role{Name: &newRoleName, ClientRole: &tr})
		if err != nil {
			klog.Error("Error when creating role", err)
			return err
		}
		klog.Infof("Role created %s", createdRoleName)
		return nil
	} else if err != nil {
		klog.Error("Error when getting user role", err)
		return err
	} else if *role.Name == newRoleName {
		klog.Infof("Role already existed %s", newRoleName)
		return nil
	}
	klog.Errorf("Error when getting role %s", newRoleName)
	return errors.New("Something went wrong when getting a role")
}

func (kcA *KcActor) deleteKcRoles(ctx context.Context, rolesToDelete []string) error {

	for _, role := range rolesToDelete {
		if err := kcA.Client.DeleteClientRole(ctx, kcA.Token.AccessToken, kcA.TargetRealm, kcA.TargetClientID, role); err != nil {
			if !strings.Contains(err.Error(), "404") {
				klog.Error("Could not delete user role", role)
				return err
			}
		}
	}
	return nil
}

func (kcA *KcActor) getUserInfo(ctx context.Context, username string) (userID *string, email *string, err error) {

	usersFound, err := kcA.Client.GetUsers(ctx, kcA.Token.AccessToken, kcA.TargetRealm, gocloak.GetUsersParams{Username: &username})
	if err != nil {
		klog.Errorf("Error when trying to find user %s", username)
		klog.Error(err)
		return nil, nil, err
	} else if len(usersFound) == 0 {
		// no existing users found, create a new one
		klog.Infof("User %s did not exists", username)
		return nil, nil, nil
	} else if len(usersFound) == 1 {
		klog.Infof("User %s already existed", username)
		return usersFound[0].ID, usersFound[0].Email, nil
	} else if len(usersFound) > 1 {
		klog.Info("Found too many users")
		return nil, nil, errors.New("Found too many users")
	}
	return nil, nil, fmt.Errorf("Error when getting user %s", username)
}

func (kcA *KcActor) createKcUser(ctx context.Context, username string, firstName string, lastName string, email string) (*string, error) {
	tr := true
	fa := false
	newUser := gocloak.User{
		Username:      &username,
		FirstName:     &firstName,
		LastName:      &lastName,
		Email:         &email,
		Enabled:       &tr,
		EmailVerified: &fa,
	}
	newUserID, err := kcA.Client.CreateUser(ctx, kcA.Token.AccessToken, kcA.TargetRealm, newUser)
	if err != nil {
		klog.Errorf("Error when creating user %s", username)
		klog.Error(err)
		return nil, err
	}
	klog.Infof("User %s created", username)
	if err = kcA.Client.ExecuteActionsEmail(ctx, kcA.Token.AccessToken, kcA.TargetRealm, gocloak.ExecuteActionsEmail{
		UserID:   &newUserID,
		Lifespan: &kcA.EmailActionsLifeSpanS,
		Actions:  &kcA.UserRequiredActions,
	}); err != nil {
		klog.Errorf("Error when sending email actions for user %s", username)
		klog.Error(err)
		return nil, err
	}

	klog.Infof("Sent verification email to user %s", username)
	return &newUserID, nil
}

func (kcA *KcActor) updateKcUser(ctx context.Context, userID string, firstName string, lastName string, email string, requireUserActions bool) error {
	tr := true
	fa := false
	updatedUser := gocloak.User{
		FirstName: &firstName,
		LastName:  &lastName,
		Email:     &email,
		Enabled:   &tr,
		ID:        &userID,
	}
	if requireUserActions {
		updatedUser.EmailVerified = &fa
	}
	err := kcA.Client.UpdateUser(ctx, kcA.Token.AccessToken, kcA.TargetRealm, updatedUser)
	if err != nil {
		klog.Errorf("Error when updating user %s %s", firstName, lastName)
		klog.Error(err)
		return err
	}
	if requireUserActions {
		if err = kcA.Client.ExecuteActionsEmail(ctx, kcA.Token.AccessToken, kcA.TargetRealm, gocloak.ExecuteActionsEmail{
			UserID:   &userID,
			Lifespan: &kcA.EmailActionsLifeSpanS,
			Actions:  &kcA.UserRequiredActions,
		}); err != nil {
			klog.Errorf("Error when sending email verification user %s %s", firstName, lastName)
			klog.Error(err)
			return err
		}
		klog.Infof("Sent user confirmation to user %s %s cause email has been updated", firstName, lastName)
	}
	return nil
}
