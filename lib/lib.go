// Copyright 2016, Duncan Gitonga



// Package lib implements method form creating and manuplation mmusic lib
package lib


import (
  "time"
  "encoding/json"
  "fmt"
  "compress/gzip"
  "io"
)

// Lib is an interface that defines methods for listing tracks
type Lib interface {
  // Return all the Tracks in a library
  Tracks() []Track

  // Track returns the track for a gine id and true if sussefull false otherwise
  Track(identifier string) (Track, bool)
}


// Track is an interface which defines methods for retrieving track metadata
type Track interface {
  // Return a string value for a given atribute
  GetString(string) string

  // Return a list of string values
  GetStrings(string) []string

  // GetInt return an int value for a given atrribute
  GetInt(string) int

  // GetTime return a time.Time value for a given attribute

  GetTime(string) time.Time
}


// library is the default internal implementation Lib which acts as the data
// source for all media tracks.
type library struct {
	trks map[string]*track
}

// Track implements Lib
func (l *library) Tracks() []Track  {
  tracks := make([]Track, 0, len(l.trks))
  for _, t := range l.trks {
    tracks = append(tracks, t)
  }
  return tracks
}

// Tracks implements Lib
func (l *library)Track(id string) (Track, bool) {
  t, ok := l.trks[id]
  return t,ok
}



// implements json.Marshal to library
func (l *library) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.trks)
}

// implements json.UnMarshal to library
func (l *library) UnmarshalJSON(b []byte) error {
	l.trks = make(map[string]*track)
	return json.Unmarshal(b, &l.trks)
}

// track is the default implementation of the Track interface.
type track struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Album       string `json:"album,omitempty"`
	AlbumArtist string `json:"albumArtist,omitempty"`
	Artist      string `json:"artist,omitempty"`
	Composer    string `json:"composer,omitempty"`
	Genre       string `json:"genre,omitempty"`
	Location    string `json:"location,omitempty"`
	Kind        string `json:"kind"`

	TotalTime   int `json:"totalTime,omitempty"`
	Year        int `json:"year,omitempty"`
	DiscNumber  int `json:"discNumber,omitempty"`
	TrackNumber int `json:"trackNumber,omitempty"`
	TrackCount  int `json:"trackCount,omitempty"`
	DiscCount   int `json:"discCount,omitempty"`
	BitRate     int `json:"bitRate,omitempty"`

	DateAdded    time.Time `json:"dateAdded,omitempty"`
	DateModified time.Time `json:"dateModified,omitempty"`
}

//  getstring id implementation
func (t *track) GetString(name string) string {
	switch name {
	case "ID":
		return t.ID
	case "Name":
		return t.Name
	case "Album":
		return t.Album
	case "AlbumArtist":
		return t.AlbumArtist
	case "Artist":
		return t.Artist
	case "Composer":
		return t.Composer
	case "Genre":
		return t.Genre
	case "Location":
		return t.Location
	case "Kind":
		return t.Kind
	}
	panic(fmt.Sprintf("unknown string field '%v'", name))
}

// DGetStrings is a function which returns the default value for a GetStrings
// attribute which is based on an existing GetString attribute.  In particular, we handle
// the case where an empty 'GetString' attribute would be "", whereas the corresponding
// 'GetStrings' method should return 'nil' and not '[]string{""}'.
func DGetStrings(t Track, f string) []string {
	v := t.GetString(f)
	if v != "" {
		return []string{v}
	}
	return nil
}

// GetStrings implements Track.
func (t *track) GetStrings(name string) []string {
	switch name {
	case "Artist", "AlbumArtist", "Composer":
		return DGetStrings(t, name)
	}
	panic(fmt.Sprintf("unknown strings field '%v", name))
}


// GetInt implements Track.
func (t *track) GetInt(name string) int {
	switch name {
	case "TotalTime":
		return t.TotalTime
	case "Year":
		return t.Year
	case "DiscNumber":
		return t.DiscNumber
	case "TrackNumber":
		return t.TrackNumber
	case "TrackCount":
		return t.TrackCount
	case "DiscCount":
		return t.DiscCount
	case "BitRate":
		return t.BitRate
	}
	panic(fmt.Sprintf("unknown int field '%v'", name))
}

// GetTime implements Track.
func (t *track) GetTime(name string) time.Time {
	switch name {
	case "DateAdded":
		return t.DateAdded
	case "DateModified":
		return t.DateModified
	}
	panic(fmt.Sprintf("unknown time field '%v'", name))
}



// Reads all the data exported by the lib and converts it to a std mmusic implementation
func  Convert(l Lib, id string) *library   {
  allTracks := l.Tracks()
	tracks := make(map[string]*track, len(allTracks))

	for _, t := range allTracks {
		identifier := t.GetString(id)
		tracks[identifier] = &track{
			// string fields
			ID:          identifier,
			Name:        t.GetString("Name"),
			Album:       t.GetString("Album"),
			AlbumArtist: t.GetString("AlbumArtist"),
			Artist:      t.GetString("Artist"),
			Composer:    t.GetString("Composer"),
			Genre:       t.GetString("Genre"),
			Location:    t.GetString("Location"),
			Kind:        t.GetString("Kind"),

			// integer fields
			TotalTime:   t.GetInt("TotalTime"),
			Year:        t.GetInt("Year"),
			DiscNumber:  t.GetInt("DiscNumber"),
			TrackNumber: t.GetInt("TrackNumber"),
			TrackCount:  t.GetInt("TrackCount"),
			DiscCount:   t.GetInt("DiscCount"),
			BitRate:     t.GetInt("BitRate"),

			// date fields
			DateAdded:    t.GetTime("DateAdded"),
			DateModified: t.GetTime("DateModified"),
		}
	}
	return &library{
		tracks,
	}
}




// WriteTo writes the Library data to the writer, currently using gzipped-JSON.
func WriteTo(l Lib, w io.Writer) error {
	gzw, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
	if err != nil {
		return err
	}
	defer gzw.Close()
	enc := json.NewEncoder(gzw)
	return enc.Encode(l)
}

//ReadFrom reads the gzipped-JSON representation of a Library.
func ReadFrom(r io.Reader) (Lib, error) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer gzr.Close()
	dec := json.NewDecoder(gzr)
	l := &library{}
	err = dec.Decode(l)
	return l, err
}
