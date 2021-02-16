/*
USAGE:
go run exportKcUsers.go\
				-path-csv=kcUsers.csv\
				-path-mail-str=kcUserMails.txt\
				-path-yaml=tenants.yaml\
				-include-roles=false\
				-kc-URL=$(KEYCLOAK_URL)\
				-kc-tenant-operator-user=$(KEYCLOAK_TENANT_OPERATOR_USER)\
				-kc-tenant-operator-psw=$(KEYCLOAK_TENANT_OPERATOR_PSW)\
				-kc-login-realm=$(KEYCLOAK_LOGIN_REALM)\
				-kc-target-realm=$(KEYCLOAK_TARGET_REALM)\
				-kc-target-client=$(KEYCLOAK_TARGET_CLIENT)

*/

package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strings"

	gocloak "github.com/Nerzal/gocloak/v7"
	"gopkg.in/yaml.v2"
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

type Tenant struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec struct {
		FirstName  string `yaml:"firstName"`
		LastName   string `yaml:"lastName"`
		Email      string `yaml:"email"`
		Workspaces []struct {
			WorkspaceRef struct {
				Name string `yaml:"name"`
			} `yaml:"workspaceRef"`
			Role string `yaml:"role"`
		} `yaml:"workspaces"`
		PublicKeys []string `yaml:"publicKeys"`
	} `yaml:"spec"`
}

func main() {

	var pathCSV string
	var pathMailStr string
	var pathYaml string
	var includeRoles bool
	var kcURL string
	var kcTenantOperatorUser string
	var kcTenantOperatorPsw string
	var kcLoginRealm string
	var kcTargetRealm string
	var kcTargetClient string

	flag.StringVar(&pathCSV, "path-csv", "", "The path for the csv file with users info.")
	flag.StringVar(&pathMailStr, "path-mail-str", "", "The path for file containing the single string with all the emails of tenants in crownlabs.")
	flag.StringVar(&pathYaml, "path-yaml", "", "The path for yml containing all the current info about a tenant to recreate on the cluster.")
	flag.BoolVar(&includeRoles, "include-roles", false, "Wether to include roles in the yml output or not")
	flag.StringVar(&kcURL, "kc-URL", "", "The URL of the keycloak client.")
	flag.StringVar(&kcTenantOperatorUser, "kc-tenant-operator-user", "", "The username of the admin account for keycloak.")
	flag.StringVar(&kcTenantOperatorPsw, "kc-tenant-operator-psw", "", "The password of the admin account for keycloak.")
	flag.StringVar(&kcLoginRealm, "kc-login-realm", "", "The realm where to login the keycloak account.")
	flag.StringVar(&kcTargetRealm, "kc-target-realm", "", "The target realm for keycloak clients and roles.")
	flag.StringVar(&kcTargetClient, "kc-target-client", "", "The target client for keycloak users and roles.")
	flag.Parse()

	if pathCSV == "" || pathMailStr == "" || pathYaml == "" ||
		kcURL == "" || kcTenantOperatorUser == "" || kcTenantOperatorPsw == "" ||
		kcLoginRealm == "" || kcTargetRealm == "" || kcTargetClient == "" {
		klog.Fatal("Some flags are not defined")
	}

	if pathCSV == pathMailStr || pathCSV == pathYaml || pathYaml == pathMailStr {
		klog.Fatal("file paths must be different from one another")
	}

	var csvFile *os.File
	csvFile, err := os.Create(pathCSV)
	if err != nil {
		klog.Fatal("Cannotcsv  create file", err)
	}
	defer csvFile.Close()
	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	tnYmlFile, err := os.Create(pathYaml)
	if err != nil {
		klog.Fatal("Cannot yml file for tenants", err)
	}
	defer tnYmlFile.Close()

	mailStr, err := os.Create(pathMailStr)
	if err != nil {
		klog.Fatal("Error when creating mail string file")
	}

	kcA, err := newKcActor(kcURL, kcTenantOperatorUser, kcTenantOperatorPsw, kcTargetRealm, kcTargetClient, kcLoginRealm)
	if err != nil {
		klog.Fatal("Error when setting up keycloak", err)
	}
	// maxUsers := 2
	maxUsers := 10000000
	usersFound, err := kcA.Client.GetUsers(context.Background(), kcA.Token.AccessToken, kcA.TargetRealm, gocloak.GetUsersParams{Max: &maxUsers})
	if err != nil {
		klog.Errorf("Error when trying to get users")
		klog.Error(err)
	} else {
		columns := []string{"UserID", "Username", "FirstName", "LastName", "Email"}
		if includeRoles {
			columns = append(columns, "Roles")
		}
		err := csvWriter.Write(columns)
		if err != nil {
			klog.Fatal("Error when writing first line of csv file")
		}
		totalUsers := len(usersFound)
		// create CSV file
		fmt.Printf("Found %d users\n", totalUsers)
		for i, user := range usersFound {
			fmt.Printf("%d / %d\n", i, totalUsers)
			tn := &Tenant{}
			tn.APIVersion = "crownlabs.polito.it/v1alpha1"
			tn.Kind = "Tenant"
			tn.Metadata.Name = *user.Username
			tn.Spec.FirstName = *user.FirstName
			tn.Spec.LastName = *user.LastName
			tn.Spec.Email = *user.Email
			userRolesString := ""

			if includeRoles {
				userRoles, errRoles := kcA.Client.GetClientRolesByUserID(context.Background(), kcA.Token.AccessToken, kcA.TargetRealm, kcA.TargetClientID, *user.ID)
				if errRoles != nil {
					klog.Fatalf("Error when getting roles of user %s", *user.Username)
				}
				userNewRoles := make(map[string]string)
				// buld new user roles
				for _, role := range userRoles {
					roleName := *role.Name
					if strings.Contains(roleName, "course-") {
						roleTrimmed := strings.TrimPrefix(*role.Name, "course-")
						if strings.HasSuffix(roleTrimmed, "admin") {
							wsName := strings.TrimSuffix(roleTrimmed, "-admin")
							fixTestCourse(&wsName)
							userNewRoles[wsName] = "manager"
						} else {
							fixTestCourse(&roleTrimmed)
							if v := userNewRoles[roleTrimmed]; v != "manager" {
								userNewRoles[roleTrimmed] = "user"
							}
						}
					}
				}

				// add workspaces
				for roleName, role := range userNewRoles {
					newWs := struct {
						WorkspaceRef struct {
							Name string `yaml:"name"`
						} `yaml:"workspaceRef"`
						Role string `yaml:"role"`
					}{}
					newWs.Role = role
					newWs.WorkspaceRef.Name = roleName
					tn.Spec.Workspaces = append(tn.Spec.Workspaces, newWs)
				}
				if len(tn.Spec.Workspaces) != 0 {
					tnYml, errYML := yaml.Marshal(tn)
					if errYML != nil {
						klog.Errorf("Error when making yml from tenant %s", tn.Metadata.Name)
					} else {
						if _, err := tnYmlFile.WriteString(fmt.Sprintf("---\n%s\n", string(tnYml))); err != nil {
							klog.Fatalf("Error when writing yml for tenant %s", *user.Username)
						}
					}
				}

				for _, v := range userRoles {
					userRolesString = fmt.Sprintf("%s $$ %s", userRolesString, *v.Name)
				}
			}
			if err = csvWriter.Write([]string{*user.ID, *user.Username, *user.FirstName, *user.LastName, *user.Email, userRolesString}); err != nil {
				klog.Fatalf("Error when writing line for user %s", *user.Username)
			}
			if _, err := mailStr.WriteString(fmt.Sprintf("%s , ", *user.Email)); err != nil {
				klog.Fatalf("Error when writing mail for tenant %s", *user.Username)
			}
		}
		if _, err := mailStr.WriteString("\n"); err != nil {
			klog.Fatalf("Error when writing last end of line for email file")
		}
	}
}

// newKcActor sets up a keycloak client with the specififed parameters and performs the first login
func newKcActor(kcURL, kcUser, kcPsw, targetRealmName, targetClient, loginRealm string) (*KcActor, error) {

	kcClient := gocloak.NewClient(kcURL)
	token, err := kcClient.LoginAdmin(context.Background(), kcUser, kcPsw, loginRealm)
	if err != nil {
		klog.Error("Unable to login as admin on keycloak", err)
		return nil, err
	}
	kcTargetClientID, err := getClientID(context.Background(), kcClient, token.AccessToken, targetRealmName, targetClient)
	if err != nil {
		klog.Errorf("Error when getting client id for %s", targetClient)
		return nil, err
	}
	return &KcActor{
		Client:         kcClient,
		Token:          token,
		TargetClientID: kcTargetClientID,
		TargetRealm:    targetRealmName,
	}, nil
}

// getClientID returns the ID of the target client given the human id, to be used with the gocloak library
func getClientID(ctx context.Context, kcClient gocloak.GoCloak, token string, realmName string, targetClient string) (string, error) {

	clients, err := kcClient.GetClients(ctx, token, realmName, gocloak.GetClientsParams{ClientID: &targetClient})
	if err != nil {
		klog.Errorf("Error when getting clientID for client %s", targetClient)
		klog.Error(err)
		return "", err
	} else if len(clients) > 1 {
		klog.Error(nil, "Error, got too many clientIDs for client %s", targetClient)
		return "", err
	} else if len(clients) < 0 {
		klog.Error(nil, "Error, no clientID for client %s", targetClient)
		return "", err

	} else {
		targetClientID := *clients[0].ID
		return targetClientID, nil
	}

}

func fixTestCourse(name *string) {
	if *name == "test-course" {
		*name = "netgroup"
	}
}
