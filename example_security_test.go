// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package loads_test

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"syscall"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/swag/loading"
)

// errForbiddenAddr is returned by the dial guard when a destination is not allowed.
var errForbiddenAddr = errors.New("blocked dial to a forbidden address")

// ExampleSpec_restrictNetwork shows how to confine remote spec loading so that a
// caller-controlled URL — or a "$ref" inside the spec — cannot reach loopback, private, or
// link-local (cloud metadata) addresses.
//
// The [net.Dialer] Control hook runs after DNS resolution and before connect, on every
// connection, so the check also covers HTTP redirects and DNS rebinding — neither of which a
// URL-string allowlist can defend against. Because the client is passed through
// [loads.WithLoadingOptions], the same guard applies to every reference resolved during
// [loads.Document.Expanded]. Here a loopback test server stands in for an internal endpoint
// that the guard must refuse to reach.
func ExampleSpec_restrictNetwork() {
	control := func(_, address string, _ syscall.RawConn) error {
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return err
		}
		addr, err := netip.ParseAddr(host)
		if err != nil {
			return err
		}
		if a := addr.Unmap(); a.IsLoopback() || a.IsPrivate() || a.IsLinkLocalUnicast() || a.IsUnspecified() {
			return errForbiddenAddr
		}

		return nil
	}

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{Control: control}).DialContext,
		},
	}

	// An internal service the application must not let untrusted input reach.
	internal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"swagger":"2.0"}`))
	}))
	defer internal.Close()

	// internal.URL is a loopback address (the untrusted URL in a real attack).
	_, err := loads.Spec(internal.URL, loads.WithLoadingOptions(loading.WithHTTPClient(client)))
	fmt.Println("blocked:", errors.Is(err, errForbiddenAddr))

	// Output:
	// blocked: true
}

// ExampleSpecRestricted shows the pre-baked restricted loader, which bundles local
// confinement and a network-restricted HTTP client. The same confinement applies to every
// "$ref" the spec resolves during expansion.
//
// Use this when the convenient, opinionated defaults fit; reach for the manual options shown
// in the other examples when you need a custom policy.
func ExampleSpecRestricted() {
	const root = "fixtures/yaml"

	// A document inside the trusted root loads normally.
	doc, err := loads.SpecRestricted("swagger/spec.yml", root)
	if err != nil {
		panic(err)
	}
	fmt.Println(doc.Host())

	// A path escaping the root is rejected.
	_, err = loads.SpecRestricted("../../../../etc/passwd", root)
	fmt.Println("escape blocked:", err != nil)

	// A remote URL pointing at an internal (loopback) address is rejected at dial time.
	internal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"swagger":"2.0"}`))
	}))
	defer internal.Close()

	_, err = loads.SpecRestricted(internal.URL, root)
	fmt.Println("network blocked:", errors.Is(err, loads.ErrForbiddenAddress))

	// Output:
	// api.example.com
	// escape blocked: true
	// network blocked: true
}

// ExampleSetRestrictedLoaders shows how to harden the package-level default in a single call,
// so every subsequent load — and every cross-package "$ref" resolution through
// spec.PathLoader — is confined, with no unconfined fallback left behind. Prefer this to
// AddLoader, which only prepends and leaves the unconfined default reachable.
func ExampleSetRestrictedLoaders() {
	loads.SetRestrictedLoaders("fixtures/yaml")
	defer loads.SetLoaders() // restore the built-in default

	doc, err := loads.Spec("swagger/spec.yml")
	if err != nil {
		panic(err)
	}
	fmt.Println(doc.Host())

	_, err = loads.Spec("../../../../etc/passwd")
	fmt.Println("escape blocked:", err != nil)

	// Output:
	// api.example.com
	// escape blocked: true
}

// ExampleSpec_restrictFilesystem shows how to confine local spec loading — and any "file://"
// reference the spec resolves — to a trusted directory.
//
// [loading.WithRoot] is built on [os.Root]: it resolves every requested path relative to the
// chosen directory and rejects anything that escapes it, whether through an absolute path,
// ".." traversal, or a symlink pointing outside. Passing it through [loads.WithLoadingOptions]
// makes the confinement apply to reference resolution as well.
func ExampleSpec_restrictFilesystem() {
	const root = "fixtures/yaml"

	// A document inside the trusted root loads normally.
	doc, err := loads.Spec("swagger/spec.yml", loads.WithLoadingOptions(loading.WithRoot(root)))
	if err != nil {
		panic(err)
	}
	fmt.Println(doc.Host())

	// An attempt to escape the root is rejected.
	_, err = loads.Spec("../../../../etc/passwd", loads.WithLoadingOptions(loading.WithRoot(root)))
	fmt.Println("escape blocked:", err != nil)

	// Output:
	// api.example.com
	// escape blocked: true
}
