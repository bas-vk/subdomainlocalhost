package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"path/filepath"
)

const (
	DAppExtension = "zip"
)

var (
	InvalidDAppArchiveError = errors.New("Invalid DApp archive") // e.g. archive isn't a zipfile
	CorruptDAppArchiveError = errors.New("Corrupt DApp archive") // e.g. invalid hash
)

// DApp instance in store
type DApp struct {
	// name of the DApp
	Name string
	// full path of DApp bundle
	Bundle string
}

type DAppCollection []*DApp

// Manages the collection of DApps
type DAppStore struct {
	dapps DAppCollection
}

// Create new DApp store with path pointing to DApp store path on disk
func NewDAppStore(path string) (*DAppStore, error) {
	store := new(DAppStore)
	err := store.loadDApps(path)
	return store, err
}

func verifyDAppBundleIntegrity(bundle *zip.ReadCloser) bool {
	return true // TODO, verify hashes and such
}

// Parse DApp bundle, filename must be the full path
// returns the parsed dapp or nil with an error in case of an error
func ParseDApp(filename string) (*DApp, error) {
	r, err := zip.OpenReader(filename)
	if err != nil {
		return nil, InvalidDAppArchiveError
	}
	defer r.Close()

	if !verifyDAppBundleIntegrity(r) {
		return nil, CorruptDAppArchiveError
	}

	// todo, maybe grab Name from manifest, for now use archive name without extension
	name := filepath.Base(filename)
	name = name[:len(name)-len(DAppExtension)-1]

	return &DApp{Name: name, Bundle: filename}, nil
}

// Load DApps from the given path
// returns collection of loaded DApps or an error
func (self *DAppStore) loadDApps(path string) error {
	pattern := fmt.Sprintf("%s%c*\\.%s", path, filepath.Separator, DAppExtension)

	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	self.dapps = make([]*DApp, 0)
	for _, filename := range files {
		dapp, err := ParseDApp(filename)
		if err != nil {
			continue
		}
		self.dapps = append(self.dapps, dapp)
	}

	return nil
}
