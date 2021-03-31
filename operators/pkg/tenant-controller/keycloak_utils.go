package tenant_controller

import (
	"context"
	"errors"
	"fmt"
	"strings"

	gocloak "github.com/Nerzal/gocloak/v7"
	"k8s.io/klog/v2"
)

// KcActor contains the needed objects and infos to use keycloak functionalities.
type KcActor struct {
	Client                gocloak.GoCloak
	Token                 *gocloak.JWT
	TargetRealm           string
	TargetClientID        string
	UserRequiredActions   []string
	EmailActionsLifeSpanS int
}

// createKcRoles takes as argument a map with each pair with the roleName as the key and its description as value.
func (kcA *KcActor) createKcRoles(ctx context.Context, rolesToCreate map[string]string) error {
	for newRoleName, newRoleDescr := range rolesToCreate {
		if err := kcA.createKcRole(ctx, newRoleName, newRoleDescr); err != nil {
			klog.Errorf("Could not create user role %s -> %s", newRoleName, err)
			return err
		}
	}
	return nil
}

func (kcA *KcActor) createKcRole(ctx context.Context, newRoleName, newRoleDescr string) error {
	tr := true
	roleAfter := gocloak.Role{Name: &newRoleName, Description: &newRoleDescr, ClientRole: &tr}

	// check if keycloak role already esists
	role, err := kcA.Client.GetClientRole(ctx, kcA.Token.AccessToken, kcA.TargetRealm, kcA.TargetClientID, newRoleName)
	if err != nil && strings.Contains(err.Error(), "Could not find role") {
		// role didn't exist
		// need to create new role
		klog.Infof("Role didn't exist %s", newRoleName)
		createdRoleName, errCreate := kcA.Client.CreateClientRole(ctx, kcA.Token.AccessToken, kcA.TargetRealm, kcA.TargetClientID, roleAfter)
		if errCreate != nil {
			klog.Errorf("Error when creating role -> %s", errCreate)
			return errCreate
		}
		klog.Infof("Role created %s", createdRoleName)
		return nil
	} else if err != nil {
		klog.Errorf("Error when getting user role -> %s", err)
		return err
	}

	if *role.Name == newRoleName {
		klog.Infof("Role already existed %s", newRoleName)
		err := kcA.Client.UpdateRole(ctx, kcA.Token.AccessToken, kcA.TargetRealm, kcA.TargetClientID, roleAfter)
		if err != nil {
			klog.Errorf("Error when creating role -> %s", err)
			return err
		}
		return nil
	}

	klog.Errorf("Error when getting role %s", newRoleName)
	return errors.New("something went wrong when getting a role")
}

func (kcA *KcActor) deleteKcRoles(ctx context.Context, rolesToDelete map[string]string) error {
	for role := range rolesToDelete {
		if err := kcA.Client.DeleteClientRole(ctx, kcA.Token.AccessToken, kcA.TargetRealm, kcA.TargetClientID, role); err != nil {
			if !strings.Contains(err.Error(), "404") {
				klog.Errorf("Could not delete user role %s -> %s", role, err)
				return err
			}
		}
	}
	return nil
}

func (kcA *KcActor) getUserInfo(ctx context.Context, username string) (userID, email *string, err error) {
	// using Exact in the GetUsersParams deosn't work cause keycloak doesn't offer the field in the API
	usersFound, err := kcA.Client.GetUsers(ctx, kcA.Token.AccessToken, kcA.TargetRealm, gocloak.GetUsersParams{Username: &username})
	if err != nil {
		klog.Errorf("Error when trying to find user %s -> %s", username, err)
		return nil, nil, err
	}

	switch len(usersFound) {
	case 0:
		// no existing users found, create a new one
		return nil, nil, nil
	case 1:
		return usersFound[0].ID, usersFound[0].Email, nil

	default:
		exactMatches := 0
		exactID := ""
		exactEmail := ""
		for _, v := range usersFound {
			if *v.Username == username {
				exactMatches++
				exactID = *v.ID
				exactEmail = *v.Email
			}
		}
		if exactMatches == 1 {
			return &exactID, &exactEmail, nil
		}
		return nil, nil, fmt.Errorf("found %d keycloak users for username %s, too many", exactMatches, username)
	}
}

func (kcA *KcActor) createKcUser(ctx context.Context, username, firstName, lastName, email string) (*string, error) {
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
		klog.Errorf("Error when creating user %s -> %s", username, err)
		return nil, err
	}
	klog.Infof("User %s created", username)
	if err = kcA.Client.ExecuteActionsEmail(ctx, kcA.Token.AccessToken, kcA.TargetRealm, gocloak.ExecuteActionsEmail{
		UserID:   &newUserID,
		Lifespan: &kcA.EmailActionsLifeSpanS,
		Actions:  &kcA.UserRequiredActions,
	}); err != nil {
		klog.Errorf("Error when sending email actions for user %s -> %s", username, err)
		return nil, err
	}

	klog.Infof("Sent verification email to user %s", username)
	return &newUserID, nil
}

func (kcA *KcActor) updateKcUser(ctx context.Context, userID, firstName, lastName, email string, requireUserActions bool) error {
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
		klog.Errorf("Error when updating user %s %s -> %s", firstName, lastName, err)
		return err
	}
	if requireUserActions {
		if err = kcA.Client.ExecuteActionsEmail(ctx, kcA.Token.AccessToken, kcA.TargetRealm, gocloak.ExecuteActionsEmail{
			UserID:   &userID,
			Lifespan: &kcA.EmailActionsLifeSpanS,
			Actions:  &kcA.UserRequiredActions,
		}); err != nil {
			klog.Errorf("Error when sending email verification user %s %s -> %s", firstName, lastName, err)
			return err
		}
		klog.Infof("Sent user confirmation to user %s %s cause email has been updated", firstName, lastName)
	}
	return nil
}

func (kcA *KcActor) updateUserRoles(ctx context.Context, roleNames []string, userID, editOnlyPrefix string) error {
	rolesToSet := make([]gocloak.Role, len(roleNames))
	// convert workspaces to actual keyloak role
	for i, roleName := range roleNames {
		// check if role exists and get roleID to use with gocloak
		gotRole, err := kcA.Client.GetClientRole(ctx, kcA.Token.AccessToken, kcA.TargetRealm, kcA.TargetClientID, roleName)
		if err != nil {
			klog.Errorf("Error when getting info on client role %s -> %s", roleName, err)
			return err
		}
		rolesToSet[i].ID = gotRole.ID
		rolesToSet[i].Name = gotRole.Name
	}
	// get current roles of user
	userCurrentRoles, err := kcA.Client.GetClientRolesByUserID(ctx, kcA.Token.AccessToken, kcA.TargetRealm, kcA.TargetClientID, userID)
	if err != nil {
		klog.Errorf("Error when getting roles of user with ID %s -> %s", userID, err)
		return err
	}
	rolesToDelete := subtractRoles(userCurrentRoles, rolesToSet, editOnlyPrefix)
	if len(rolesToDelete) > 0 {
		// this is idempotent
		err = kcA.Client.DeleteClientRoleFromUser(ctx, kcA.Token.AccessToken, kcA.TargetRealm, kcA.TargetClientID, userID, rolesToDelete)
		if err != nil {
			klog.Errorf("Error when removing user roles to user with ID %s -> %s", userID, err)
			return err
		}
	}
	// // this is idempotent
	err = kcA.Client.AddClientRoleToUser(ctx, kcA.Token.AccessToken, kcA.TargetRealm, kcA.TargetClientID, userID, rolesToSet)
	if err != nil {
		klog.Errorf("Error when adding user roles to user with ID %s -> %s", userID, err)
		return err
	}
	return nil
}

func subtractRoles(a []*gocloak.Role, b []gocloak.Role, subtractOnlyPrefix string) []gocloak.Role {
	var res []gocloak.Role
	// temporary map to hold values of b for faster subtraction in sacrifice of memory
	tempMap := make(map[string]bool)

	for _, role := range b {
		tempMap[*role.Name] = true
	}

	for _, role := range a {
		if strings.HasPrefix(*role.Name, subtractOnlyPrefix) {
			if _, ok := tempMap[*role.Name]; !ok {
				res = append(res, *role)
			}
		}
	}

	return res
}
