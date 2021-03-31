package bastion_controller

import (
	"errors"
	"os"
	"strings"

	"k8s.io/klog/v2"
)

func closeFile(f *os.File) {
	err := f.Close()
	if err != nil {
		klog.Errorf("unable to close the file authorized_keys: %v", err)
	}
}

func removeKey(s []string, i int) []string {
	if i >= len(s) {
		return s
	}
	s[i] = s[len(s)-1]

	return s[:len(s)-1]
}

// AuthorizedKeysEntry is a structure containing the three different fields
// of an entry of the .ssh/authorized_keys file.
type AuthorizedKeysEntry struct {
	Algo, Key, ID string
}

// Decompose converts a string into an AuthorizedKeysEntry object.
func Decompose(entry string) (AuthorizedKeysEntry, error) {
	entryComponents := strings.Split(entry, string(" "))
	if len(entryComponents) == 3 {
		return AuthorizedKeysEntry{
			Algo: entryComponents[0],
			Key:  entryComponents[1],
			ID:   entryComponents[2],
		}, nil
	}

	return AuthorizedKeysEntry{}, errors.New("invalid entry")
}

// Create converts a string and an id into an AuthorizedKeysEntry object.
func Create(entry, id string) (AuthorizedKeysEntry, error) {
	entryComponents := strings.Split(entry, string(" "))
	if len(entryComponents) == 3 || len(entryComponents) == 2 {
		return AuthorizedKeysEntry{
			Algo: entryComponents[0],
			Key:  entryComponents[1],
			ID:   id,
		}, nil
	}

	return AuthorizedKeysEntry{}, errors.New("invalid entry")
}

// Compose an AuthorizedKeysEntry object into a string.
func (e *AuthorizedKeysEntry) Compose() string {
	return e.Algo + " " + e.Key + " " + e.ID
}

func decomposeAndPurgeEntries(keys []string, tenantID string) []string {
	indexesList := []int{}
	for i, key := range keys {
		entry, err := Decompose(key)
		if err != nil {
			continue
		}
		if entry.ID == tenantID {
			indexesList = append(indexesList, i)
		}
	}

	// we have to iterate in reverse order, otherwise the last indexes could become out of range
	for k := len(indexesList) - 1; k >= 0; k-- {
		keys = removeKey(keys, indexesList[k])
	}
	return keys
}

func composeAndMarkEntries(keys, tenantKeys []string, tenantID string) []string {
	for i := range tenantKeys {
		entry, err := Create(tenantKeys[i], tenantID)
		if err != nil {
			continue
		}
		keys = append(keys, entry.Compose())
	}
	return keys
}
