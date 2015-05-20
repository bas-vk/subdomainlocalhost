package main

import (
	"archive/zip"
	"errors"
	"net/http"
	"strings"

	"io"

	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/zipfs"
)

var (
	DuplicateDAppError = errors.New("Duplicated DApp")
)

// Represents a DApp subdomain http service which handles requests on
// http://{DAppInstance.Name}.localhost
type DAppService struct {
	// name of the DApp
	Name string
	// full path of DApp bundle
	Bundle string
	// subdomain where this DApp is served from
	Subdomain string
	// DApp virtual filesystem
	vfs vfs.FileSystem
}

// Handle http request
func (self *DAppService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.RequestURI()

	if fileReader, err := self.vfs.Open(uri); err == nil {
		defer fileReader.Close()
		_, err = io.Copy(w, fileReader)
        if err != nil {
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
}

// Request multiplexer which calls HttpServe on the correct DAppInstance
type DAppStoreService struct {
	handlers map[string]http.Handler
}

// Create a new http multiplexer with for each DApp in the store a separate subdomain
// e.g. http://wallet.localhost, http://exchange.localhost
func NewDAppMux(store *DAppStore) (*DAppStoreService, error) {
	svc := &DAppStoreService{handlers: make(map[string]http.Handler)}
	for _, dapp := range store.dapps {
		if err := svc.add(dapp); err != nil {
			return nil, err
		}
	}

	return svc, nil
}

// Create mux for DApp and create handlers
func (self *DAppStoreService) add(dapp *DApp) error {
	subdomain := strings.ToLower(dapp.Name)
	if _, found := self.handlers[subdomain]; found {
		return DuplicateDAppError
	}

    // serve DApp pages direct from the zip bundle
    // (must be replaces when DApps are served from swarm)
	r, err := zip.OpenReader(dapp.Bundle)
	if err != nil {
		return err
	}

	svc := &DAppService{
		Name:      dapp.Name,
		Subdomain: strings.ToLower(dapp.Name),
		Bundle:    dapp.Bundle,
		vfs:       zipfs.New(r, dapp.Bundle),
	}
	self.handlers[svc.Subdomain] = svc

	return nil
}

// 1. Determine for which subdomain this request belongs to
// 2. Lookup the corresponding subdomain handler
// 3. call handler, or return an error if not found
func (self *DAppStoreService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	host = strings.TrimSpace(host)
	hostParts := strings.Split(host, ".")

	if len(hostParts) == 2 && strings.HasPrefix(hostParts[1], "localhost:") {
		if mux, found := self.handlers[strings.ToLower(hostParts[0])]; found {
			mux.ServeHTTP(w, r)
			return
		}
	}

	http.Error(w, "Not found", http.StatusNotFound)
}
