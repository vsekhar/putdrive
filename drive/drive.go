package drive

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/vsekhar/putdrive/flags"
	"github.com/vsekhar/putdrive/credentials"

	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/google-api-go-client/drive/v2"
)

var config = &oauth.Config{
	ClientId:     credentials.DriveClientId,
	ClientSecret: credentials.DriveClientSecret,
	Scope:        drive.DriveFileScope,
	RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
	AuthURL:      "https://accounts.google.com/o/oauth2/auth",
	TokenURL:     "https://accounts.google.com/o/oauth2/token",
}

const DriveFolderType = "application/vnd.google-apps.folder"

// Get auth from user (requires manual entry)
func oauthTransport() *oauth.Transport {
	transport := &oauth.Transport {
		Config:    config,
		Transport: http.DefaultTransport,
	}
	if credentials.DriveToken != nil {
		// use pre-configured token
		log.Printf("using pre-configured OAUTH2 token")
		transport.Token = credentials.DriveToken
	} else {
		// get a token
		log.Printf("No pre-configured OAUTH2 token")
		authurl := config.AuthCodeURL("state")
		log.Printf("Go to the following link in your browser: %v\n", authurl)
		log.Printf("Enter verification code: ")
		var code string
		fmt.Scanln(&code)
		_, err := transport.Exchange(code)
		if err != nil {
			log.Fatalf("exchanging the code: %v\n", err)
		}
	}
	// for debugging
	if *flags.ShowTokens {
		fmt.Printf("AccessToken: %s\n", transport.Token.AccessToken)
		fmt.Printf("RefreshToken: %s\n", transport.Token.RefreshToken)
		fmt.Printf("Expiry: %s\n", transport.Token.Expiry)
	}
	return transport
}

func driveClient(t *oauth.Transport) *drive.Service {
	svc, err := drive.New(t.Client())
	if err != nil {
		log.Fatalf("creating drive client: %v", err)
	}
	return svc
}

// Wrapper for convenience functions
type Entry struct {
	dsvc *drive.Service
	f *drive.File
}

func service() *drive.Service {
	t := oauthTransport()
	dsvc, err := drive.New(t.Client())
	if err != nil {
		log.Fatalf("creating drive client: %v", err)
	}
	return dsvc
}

func Folder(f *drive.File) *Entry {
	return &Entry{
		dsvc: service(),
		f: f,
	}
}

func Root() *Entry {
	return &Entry{
		dsvc: service(),
	}
}

func (d *Entry) Size() int64 {
	return d.f.FileSize
}

func (d *Entry) ParentReference() []*drive.ParentReference {
	if d.f != nil {
		return []*drive.ParentReference{&drive.ParentReference{Id: d.f.Id}}
	}
	return []*drive.ParentReference{}
}

func (d *Entry) CreateFile(name string, data io.Reader) *Entry {
	return d.createImpl(name, data, false)
}

func (d *Entry) CreateFolder(name string) *Entry {
	return d.createImpl(name, nil, true)
}

func (d *Entry) createImpl(name string, data io.Reader, folder bool) *Entry {
	file := &drive.File{
		Title: name,
		Parents: d.ParentReference(),
	}
	if folder {
		file.MimeType = "application/vnd.google-apps.folder"
	}
	fr := d.dsvc.Files.Insert(file)
	if data != nil {
		fr = fr.Media(data)
	}
	newf, err := fr.Do()
	if err != nil {
		log.Fatalf("creating %+v: %v", file, err)
	}
	return &Entry{d.dsvc, newf}
}
