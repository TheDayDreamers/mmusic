// Copyright 2016, Duncan Gitonga


// Package walk implements path walker that reads all the audio file based on tags see(github.com/dhowden/tag)
package walk

import (
  "fmt"
  "strings"
  "path/filepath"
  "time"
  "os"
  "crypto/sha1"
  "log"
  "sync"


  "github.com/dhowden/tag"
  "github.com/Duncodes/mmusic/lib"
)



var workers = 4

type pathTrack struct {
	path  string
	track *track
}

var FileExtentions = []string {".mp3",".m4a",".ogg",".flac"}

// validFiles take an in chan of filepaths and return a channel of validFiles in the given path
func validFiles(in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		for path := range in {
			ext := strings.ToLower(filepath.Ext(filepath.Base(path)))
			for _, x := range FileExtentions {
				if ext == x {
					out <- path
					break
				}
			}
		}
		close(out)
	}()
	return out
}

// NewLibrary constract a lib.Lib by walking through the directory tree under
// the given path
func NewLibrary(path string) lib.Lib {
  trackCh := make(chan pathTrack)
  errCh := make(chan error)
  files := validFiles(walk(path))
  go func() {
    for err := range errCh {
      // BUG dont write error to the log os
      log.Println(err)
  }
}()
process := func(files <-chan string) {
  for p := range files {
    t, err := processPath(p)
    if err != nil {
      errCh <- fmt.Errorf("error processing '%v': %v", p, err)
      continue
    }
    trackCh <- pathTrack{p, t}
  }
}
wg := &sync.WaitGroup{}
wg.Add(workers)
for i := 0; i < workers; i++ {
  go func() {
    process(files)
    wg.Done()
  }()
}

go func() {
  wg.Wait()
  close(errCh)
  close(trackCh)
}()

tracks := make(map[string]*track)
for pt := range trackCh {
  tracks[pt.path] = pt.track
}

return &library{
  tracks: tracks,
}
}

// library is an implementation of lib.lib.
type library struct {
	tracks map[string]*track
}

// Track implements lib.Lib.
func (l *library) Track(id string) (lib.Track, bool) {
	t, ok := l.tracks[id]
	return t, ok
}

// Tracks implements lib.Lib.
func (l *library) Tracks() []lib.Track {
	tracks := make([]lib.Track, 0, len(l.tracks))
	for _, t := range l.tracks {
		tracks = append(tracks, t)
	}
	return tracks
}

// track is a wrapper around tag.Metadata which implements lib.Track
type track struct {
	tag.Metadata
	Location    string
	FileInfo    os.FileInfo
	CreatedTime time.Time
}

// GetString implements lib.Track.
func (m *track) GetString(name string) string {
	switch name {
	case "Name":
		title := m.Title()
		if title == "" {
			fileName := m.FileInfo.Name()
			ext := filepath.Ext(fileName)
			title = strings.TrimSuffix(fileName, ext)
		}
		return title
	case "Album":
		return m.Album()
	case "Artist":
		return m.Artist()
	case "AlbumArtist":
		return m.AlbumArtist()
	case "Composer":
		return m.Composer()
	case "Genre":
		return m.Genre()
	case "Location":
		return m.Location
	case "Kind":
		return kind(m.FileType()).String()
	case "ID":
		sum := sha1.Sum([]byte(m.Location))
		return string(fmt.Sprintf("%x", sum))
	}
	return ""
}

type kind tag.FileType

func (k kind) String() string {
	switch k {
	case tag.MP3:
		return "MPEG audio file"
	case tag.M4A:
		return "AAC audio file"
	case tag.ALAC: // FIXME: tag doesn't actually detect this at the moment.
		return "Apple Lossless audio file"
	case tag.FLAC:
		return "FLAC audio file"
	case tag.OGG:
		return "OGG audio file"
	}
	return ""
}

// GetStrings implements Lib.Track.
func (m *track) GetStrings(name string) []string {
	switch name {
	case "Artist", "AlbumArtist", "Composer":
		return lib.DGetStrings(m, name)
	}
	return nil
}

// GetInt implements Lib.Track.
func (m *track) GetInt(name string) int {
	switch name {
	case "Year":
		return m.Year()
	case "TrackNumber":
		x, _ := m.Track()
		return x
	case "TrackCount":
		_, n := m.Track()
		return n
	case "DiscNumber":
		x, _ := m.Disc()
		return x
	case "DiscCount":
		_, n := m.Disc()
		return n
	}
	return 0
}

// GetTime implements lib.Truck
func (m *track) GetTime(name string) time.Time {
	switch name {
	case "DateModified":
		return m.FileInfo.ModTime()
	case "DateAdded":
		return m.CreatedTime
	}
	return time.Time{}
}

func walk(root string) <-chan string {
	ch := make(chan string)
	fn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ch <- path
		return nil
	}

	go func() {
		err := filepath.Walk(root, fn)
		if err != nil {
			log.Println(err)
		}
		close(ch)
	}()
	return ch
}

func processPath(path string) (*track, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil {
		return nil, err
	}

	fileInfo, err := f.Stat()
	if err != nil {
		return nil, err
	}

	createdTime, err := getCreatedTime(path)
	if err != nil {
		return nil, err
	}

	return &track{
		Metadata:    m,
		Location:    path,
		FileInfo:    fileInfo,
		CreatedTime: createdTime,
	}, nil
}
