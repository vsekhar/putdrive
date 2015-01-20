package putio

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/vsekhar/govtil/log"
)

var baseUrl = url.URL{
	Scheme: "https",
	Host: "api.put.io",
	Path: "/v2",
}

const folderType = "application/x-directory"

// An authenticated connection to put.io
type PutIOService struct {
	OauthToken string
	Client http.Client
}

// An entry represents a file or folder on put.io
type Entry struct {
	Name string `json:"name"`
	ContentType string `json:"content_type"`
	Id int
	Parent int `json:"parent_id"`
	Size int64 `json:"size"`
	svc *PutIOService
	path string
}

// Create a new connection with the given OAUTH token
func NewPutIOService(token string) *PutIOService {
	return &PutIOService{token, http.Client{}}
}

func (p *PutIOService) EntryById(id int) *Entry {
	resp := p.get("/files/"+fmt.Sprint(id), nil, nil)

	var ir struct {
		File Entry `json:"file"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp).Decode(&ir); err != nil {
		log.Fatalf("decoding id JSON: %v", err)
	}
	if ir.Status != "OK" {
		log.Errorf("bad status for put.io id %d", id)
		return nil
	}
	ir.File.svc = p
	return &ir.File
}

func (p *PutIOService) get(path string, v url.Values, h http.Header) io.ReadCloser {
	// add query values to path
	u := baseUrl
	u.Path += path
	query := make(url.Values)
	query.Add("oauth_token", p.OauthToken)
	for k, s := range v {
		for _, val := range s {
			query.Add(k, val)
		}
	}
	u.RawQuery = query.Encode()
	return p.do("GET", u.String(), h, "", nil)
}

func (p *PutIOService) postValues(path string, v url.Values, h http.Header) io.ReadCloser {
	u := baseUrl
	u.Path += path
	q := make(url.Values)
	q.Add("oauth_token", p.OauthToken)
	u.RawQuery = q.Encode()
	return p.post(u.String(), h, "application/x-www-form-urlencoded", strings.NewReader(v.Encode()))
}

func (p *PutIOService) post(path string, h http.Header, bodytype string, body io.Reader) io.ReadCloser {
	return p.do("POST", path, h, bodytype, body)
}

func (p *PutIOService) do(method string, path string, h http.Header, bodytype string, body io.Reader) io.ReadCloser {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		log.Fatalf("creating request: %v", err)
	}
	if h != nil {
		req.Header = h
	}
	if body != nil {
		req.Header.Set("Content-Type", bodytype)
	}

	// This seems to be ok even for binary downloads
	req.Header.Set("Accept", "application/json")

	log.Debugf("Request: %v", req)
	resp, err := p.Client.Do(req)
	if err != nil {
		log.Debugf("making request: %v", err)
		log.Fatalf("response: %s", resp)
	}
	return resp.Body
}

func (s *PutIOService) Root() *Entry {
	ret := s.EntryById(0)
	if ret == nil {
		log.Fatalf("cannot get root")
	}
	ret.svc = s
	return ret
}

func (e *Entry) Path() string {
	if e.path != "" {
		return e.path
	}
	if e.Id == 0 {
		return ""
	}
	p := e.svc.EntryById(e.Parent)
	if p == nil {
		log.Fatalf("cannot get parent")
	}
	e.path = p.Path() + "/" + e.Name
	return e.path
}

func (e *Entry) IsFolder() bool {
	return e.ContentType == folderType
}

func (e *Entry) List() []*Entry {
	if !e.IsFolder() {
		log.Fatalf("tried to list non-folder")
	}
	v := url.Values{}
	v.Set("parent_id", fmt.Sprint(e.Id))
	resp := e.svc.get("/files/list", v, nil)

	var lr struct {
		Files []*Entry `json:"files"`
	}
	if err := json.NewDecoder(resp).Decode(&lr); err != nil {
		log.Fatalf("decoding JSON: %v", err)
	}

	for i, _ := range lr.Files {
		lr.Files[i].svc = e.svc
	}

	return lr.Files
}

func (e *Entry) Download() io.ReadCloser {
	return e.DownloadRange(0,0)
}

func (e *Entry) DownloadRange(start, end int) io.ReadCloser {
	if e.IsFolder() {
		log.Fatalf("tried to download a folder")
	}
	var h map[string][]string
	if start != 0 && end != 0 {
		h = make(map[string][]string)
		h["Range"] = []string{"bytes=" + fmt.Sprint(start) + "-" + fmt.Sprint(end)}
	}
	return e.svc.get("/files/"+fmt.Sprint(e.Id)+"/download", nil, h)
}

// Using this Entry struct after Delete is undefined
func (e *Entry) Delete() error {
	log.Printf("deleting %s (%d)", e.Path(), e.Id)
	v := url.Values{"file_ids": {fmt.Sprint(e.Id)}}
	resp := e.svc.postValues("/files/delete", v, nil)
	var dr struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp).Decode(&dr); err != nil {
		log.Fatalf("decoding id JSON: %v", err)
	}
	log.Debugf("delete response: %+v", dr)
	if dr.Status != "OK" {
		return fmt.Errorf("bad status for put.io delete(%d): %s", e.Id, dr.Status)
	}
	return nil
}
