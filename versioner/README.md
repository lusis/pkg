# versioner

Versioner is a helper library for versioning types

The primary use case is for versioning types that correspond to api requests.

By satisfying the versioner interface, you can perform checks in your api client libraries before calling an api.

## Usage

- Satisfy the `Versioner` interface in your structs
- Call `CheckSupportedVersion` at the start of any functions that could use those types

e.g.

```go
type Foo struct{}
func (f *Foo) MinVersion() versioner.Version {
    return versioner.MustParse("1.2.x")
}
func (f *Foo) MaxVersion() versioner.Version {
    return versioner.MustParse("2.0.x")
}
func (f *Foo) Deprecated() bool {
    return false
}

func (f *Foo) DoFooThing() (*Bar, error) {
    if err := versioner.CheckSupportedVersions(f, "1.3.0"); err != nil {
        return nil, err
    }
}
```

Where you get the version to pass to `CheckSupportedVersions` is up to you. My client objects all have a config struct field that contains the API version:

```go
type ClientConfig struct {
    BaseURL    string
    Username   string
    Password   string
    Token      string
    AuthMethod string
    VerifySSL  bool
    Debug      bool
    APIVersion Version
    HTTPClient *http.Client
}

// Client is an artifactory client
type Client struct {
    Config     *ClientConfig
    HTTPClient *http.Client
}
```

## Background

I like writing API client libraries for various tools I use. The two big ones are rundeck and artifactory.
I struggled for a long time to figure out how to maintain changing the code as the upstream API changed
in a way that was ergonomic.

For the longest time, I ran with directories like `artifactory.v51` and `rundeck.v17` but this got painful.
Even just copying over the previous version and doing a big ole search and replace left something to be desired.

Additionally, I shipped cli tooling for interacting with those services and by changing the version the cli tooling used, I risked breaking it for someone running an older version of those servers.

As part of thinking about this I came to the idea after looking at multiple different repositories of breaking out my request/response parsing structs into a discrete package. This made things SUPER flexible.

But I still struggled with how to handle version changes between the API for my tooling and end users.

My ultimate goal was to downgrade successfully for older users.

## First attempt

I decided to start versioning my structs by having them satisfy an interface:

The following is from the current implementation in [go-rundeck](https://github.com/lusis/go-rundeck/blob/master/pkg/rundeck/responses/versioned_response.go):

```go
package responses

// VersionedResponse is an interface for a Rundeck Response that supports versioning information
type VersionedResponse interface {
    minVersion() int
    maxVersion() int
    deprecated() bool
}

// AbsoluteMinimumVersion is the absolute minimum version this library will support
// We set this to `14` as that was the first version of the rundeck API to support JSON
const AbsoluteMinimumVersion = 14

// CurrentVersion is the current version of the API that this library is tested against
const CurrentVersion = 21

// GetMinVersionFor gets the minimum api version required for a response
func GetMinVersionFor(a VersionedResponse) int { return a.minVersion() }

// GetMaxVersionFor gets the maximum api version required for a response
func GetMaxVersionFor(a VersionedResponse) int { return a.maxVersion() }

// IsDeprecated indicates if a response is deprecated or not
func IsDeprecated(a VersionedResponse) bool { return a.deprecated() }

// GenericVersionedResponse is for version checking
// Some operations don't have a response (think DELETE or PUT)
// but we still want to do a version check on ALL functions anyway
// This response simply responds to that
type GenericVersionedResponse struct{}

func (g GenericVersionedResponse) minVersion() int  { return AbsoluteMinimumVersion }
func (g GenericVersionedResponse) maxVersion() int  { return CurrentVersion }
func (g GenericVersionedResponse) deprecated() bool { return false }

```

In the higher level package which handles actually MAKING api calls, I now do this:

```go
// ListUsers returns all rundeck users
// http://rundeck.org/docs/api/index.html#list-users
func (c *Client) ListUsers() (Users, error) {
    if err := c.checkRequiredAPIVersion(responses.ListUsersResponse{}); err != nil {
        return nil, err
    }
    # actual work
}
```

This worked well since the rundeck api uses the version of the API in the request URI:

- `/api/21/user/list/`
- `/api/14/system/info`

Now if a user is using an older version of the Rundeck server, they can override the `APIVersion` in the client config. While they won't be able to call `ListUsers`, the can call `GetSystemInfo`.

## Next iteration

I recently started to port some of this logic over to my Artifactory client as part of rewriting it.

Artifactory is a little different. The API version is NOT a part of the URI and the versions I need to match are `X.Y.Z`. This meant I had to change my process a little bit.

I've worked with [go-version](https://github.com/hashicorp/go-version) in the past so I pulled it into my Artifactory client as I started versioning my structs ([full file here](https://github.com/lusis/go-artifactory/blob/single-dir/pkg/artifactory/responses/versioned_response.go))

```go
package responses

import (
    gover "github.com/hashicorp/go-version"
)

// Version is a self-contained go-version Version
type Version struct {
    *gover.Version
}

// VersionedResponse is an interface for a Rundeck Response that supports versioning information
type VersionedResponse interface {
    minVersion() Version
    maxVersion() Version
    deprecated() bool
}
```

It's a bit different since my versions are now strings. To keep the function signatures similar and simple, I created a panicing version of `NewVersion` that lines up with other patterns in the Go stdlib, `MustParse`:

```go
// versionMustParse is a panicing version of NewVersion
func versionMustParse(v string) Version {
    ver, err := gover.NewVersion(v)
    if err != nil {
        panic("cannot parse version")
    }
    return Version{ver}
}
```

Now I have the same functionality as my rundeck library for versioning.

## This package

I found myself having to duplicate some of the logic in a few places (`VersionedResponse` and `VersionedRequest`) so I extracted everything out into its own package which is what this is.