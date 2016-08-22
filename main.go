
// The main package
package main

import (
       "log"
        // "github.com/dhowden/tag"
        "github.com/Duncodes/mmusic/walk"
        // "os"
        "flag"
        "fmt"
        "github.com/Duncodes/mmusic/lib"

)
// set run flag
var strFlag = flag.String("path","~/","path")

func init() {
	 flag.StringVar(strFlag, "s", "", "Description")
}



type TestLib interface {

}
// NOTE here is a metadata interface
// type Metadata interface {
//     Format() Format
//     FileType() FileType

//     Title() string
//     Album() string
//     Artist() string
//     AlbumArtist() string
//     Composer() string
//     Genre() string
//     Year() int

//     Track() (int, int) // Number, Total
//     Disc() (int, int) // Number, Total

//     Picture() *Picture // Artwork
//     Lyrics() string

//     Raw() map[string]interface{} // NB: raw tag names are not consistent across formats.
// }


func main() {
	flag.Parse()
	root_dir := *strFlag
	// f ,_:=os.Open(root_dir)
    // _ ,err := tag.ReadFrom(f)
    // if err !=nil{
    	// log.Fatal(err)
    // }
    TestLib := walk.NewLibrary(root_dir)
    json := TestLib.MarshalJSON()
    fmt.Println("%v",TestLib)
    log.Print(walk.FileExtentions[0])
    // log.Print(m.Format())
    // log.Print(m.Title())
    // log.Print(m.FileType())
    // log.Print(m.Artist())
    // log.Print(m.Year())
}
