package main

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/vsekhar/putdrive/credentials"
	"github.com/vsekhar/putdrive/drive"
	"github.com/vsekhar/putdrive/flags"
	"github.com/vsekhar/putdrive/putio"
)

func main() {
	// Services
	dsvc := drive.NewDriveService(credentials.DriveAccessToken,
                                  credentials.DriveRefreshToken,
		                          credentials.DriveExpiry)
	psvc := putio.NewPutIOService(credentials.PutIOToken)

	// Create a folder with the current time and work within it
	df := dsvc.ParentFolder(credentials.DriveParentFolderId)
	if *flags.Copy {
		t := time.Now().Format(time.RFC3339)
		log.Printf("Storing in drive: '%s'", t)
		df = df.CreateFolder(t)
	}

	if *flags.ItemIds == "" {
		log.Printf("Syncing from root")
		entry := psvc.Root()
		if err := Walk(entry, df); err != nil {
			log.Fatalf("error walking %s: %s", entry.Path(), err)
		}
	} else {
		items := strings.Split(*flags.ItemIds, ",")
		for i := range items {
			items[i] = strings.TrimSpace(items[i])
		}
		log.Printf("Syncing items: %v", items)
		for _, id := range items {
			id_i, err := strconv.Atoi(id)
			if err != nil {
				log.Printf("Bad put.io file/folder id (%d): %v", id, err)
				continue
			}
			entry := psvc.EntryById(id_i)
			if entry == nil {
				// not found or error
				continue
			}
			log.Printf("Syncing %s (%d)", entry.Path(), entry.Id)
			if err := Walk(entry, df); err != nil {
				log.Fatalf("error walking %s: %s", entry.Path(), err)
			}
		}
	}
}

func Walk(p *putio.Entry, d *drive.Entry) error {
	if p.IsFolder() {
		// recurse (folder)
		log.Printf("Entering folder: %s", p.Path())
		if *flags.Copy {
			d = d.CreateFolder(p.Name)
		}
		pcs := p.List()
		for _, newp := range pcs {
			if err := Walk(newp, d); err != nil {
				return err
			}
		}
		if *flags.Delete {
			if err := p.Delete(); err != nil {
				return err
			}
		}
	} else {
		// base (file)
		if *flags.Copy {
			var newe *drive.Entry
			func() {
				dr := p.Download()
				defer dr.Close()
				newe = d.CreateFile(p.Name, dr)
			}()
			if newe.Size() != p.Size {
				log.Fatalf("Error copying '%s', %d bytes after move (should be %d bytes)", p.Path(), newe.Size(), p.Size)
			}
			log.Printf("Copied file: %s (%d bytes)", p.Path(), p.Size)
		}
		if *flags.Delete {
			if err := p.Delete(); err != nil {
				return err
			}
		}
	}
	return nil
}
