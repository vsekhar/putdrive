// Populate this, remove build ignore flag and save as 'credentials.go' before building

// +build ignore

package credentials

import (
	"time"

	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/google-api-go-client/drive/v2"
)

// Get these by registering an app at https://console.developers.google.com/ and
// creating an API client
const DriveClientId = "YOUR_CLIENT_ID"
const DriveClientSecret = "YOUR_CLIENT_SECRET"

// Get these by running with --show-tokens and noting the values
var DriveToken = &oauth.Token {
	AccessToken: "CACHED_ACCESS_TOKEN",
	RefreshToken: "CACHED_REFRESH_TOKEN",
	Expiry: time.Date(2014, time.November, 16, 21, 38, 19, 447586, time.FixedZone("PST", -8*60*60)),
}

// Get this by looking at the URL of the Drive folder you want to put files in
// or "" for root folder
var DriveParentFolder = &drive.File{
	Id: "DRIVE_FOLDER_ID",
}

// Get this by creating a client at put.io
const PutIOToken = "PUTIOTOKEN"
