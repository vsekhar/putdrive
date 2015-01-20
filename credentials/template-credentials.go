// Populate this, remove build ignore flag and save as 'credentials.go' before building

// +build ignore

package credentials

import "time"

// Get these by registering an app at https://console.developers.google.com/ and
// creating an API client
const DriveClientId = "YOUR_CLIENT_ID"
const DriveClientSecret = "YOUR_CLIENT_SECRET"

// Get these by running with --show-tokens and noting the values
const DriveAccessToken = "CACHED_ACCESS_TOKEN"
const DriveRefreshToken = "CACHED_REFRESH_TOKEN"
var DriveExpiry = time.Date(2014, time.November, 16, 21, 38, 19, 447586, time.FixedZone("PST", -8*60*60))

// Get this by looking at the URL of the Drive folder you want to put files in
// or "" for root folder
const DriveParentFolderId = "DRIVE_FOLDER_ID"
}

// Get this by creating a client at put.io
const PutIOToken = "PUTIOTOKEN"
