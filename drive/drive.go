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

func oauthTransport(token *oauth.Token) *oauth.Transport {
	transport := &oauth.Transport {
		Config:    config,
		Transport: http.DefaultTransport,
	}
	if token == nil || *flags.NewTokens {
		// get a token
		log.Printf("Getting a new token")
		authurl := config.AuthCodeURL("state")
		log.Printf("Go to the following link in your browser: %v\n", authurl)
		log.Printf("Enter verification code: ")
		var code string
		fmt.Scanln(&code)
		_, err := transport.Exchange(code)
		if err != nil {
			log.Fatalf("exchanging the code: %v\n", err)
		}
	} else {
		// use pre-configured token
		log.Printf("using pre-configured OAUTH2 token")
		transport.Token = token
	}
	// for debugging
	if *flags.ShowTokens {
		fmt.Printf("AccessToken: %s\n", transport.Token.AccessToken)
		fmt.Printf("RefreshToken: %s\n", transport.Token.RefreshToken)
		fmt.Printf("Expiry: %s\n", transport.Token.Expiry)
	}
	return transport
}

type Service struct {
	gsvc *drive.Service
}

// File or folder
type Entry struct {
	svc *Service
	f *drive.File
}

func NewDriveService(token *oauth.Token) *Service {
	t := oauthTransport(token)
	gsvc, err := drive.New(t.Client())
	if err != nil {
		log.Fatalf("creating drive client: %v", err)
	}
	return &Service{gsvc}
}

// A parent folder is a special entry that can only be used to create child
// entries. It is not itself accessible. Calling anything other than Create*()
// on the Entry returned by ParentFolder is undefined.
//
// This is required to allow writing to any parent folders while holding only
// Drive.File permission (and not full Drive permission).
func (s *Service) ParentFolder(id string) *Entry {
	return &Entry{
		svc: s,
		f: &drive.File{Id: id},
	}
}

func (s *Service) Item(id string) *Entry {
	f, err := s.gsvc.Files.Get(id).Do()
	if err != nil {
		log.Fatalf("failed to get Drive file '%s': %s", id, err)
	}
	return &Entry{
		svc: s,
		f: f,
	}
}

func (d *Entry) Size() int64 {
	return d.f.FileSize
}

// parent reference to this entry (slice because files can have many)
func (d *Entry) parentReference() []*drive.ParentReference {
	return []*drive.ParentReference{&drive.ParentReference{Id: d.f.Id}}
}

func (d *Entry) CreateFile(name string, data io.Reader) *Entry {
	return d.createImpl(name, data, false)
}

func (d *Entry) CreateFolder(name string) *Entry {
	return d.createImpl(name, nil, true)
}

func (d *Entry) createImpl(name string, data io.Reader, folder bool) *Entry {
	file := &drive.File{}
	file.Title = name
	file.Parents = append(file.Parents, &drive.ParentReference{Id: d.f.Id})
	if folder {
		file.MimeType = "application/vnd.google-apps.folder"
	}
	fr := d.svc.gsvc.Files.Insert(file)
	if data != nil {
		fr = fr.Media(data)
	}
	newf, err := fr.Do()
	if err != nil {
		log.Fatalf("creating %+v: %v", file, err)
	}
	return &Entry{d.svc, newf}
}

func (d *Entry) createMultipartFile(name string, data io.Reader) *Entry {
	file := &drive.File{}
	file.Title = name
	file.Parents = append(file.Parents, &drive.ParentReference{Id: d.f.Id})
	return nil
}
