package main

import (
	"log"

	"github.com/HardDie/fsentry"
)

// Some type with useful information that we will store inside the entry.
type Data struct {
	Title string
	Value int
}

func main() {
	// Initializing the fsentry repository.
	db := fsentry.NewFSEntry("test",
		fsentry.WithPretty(), // If you want to store json metadata files in a pretty format.
	)

	// Check if a repository folder has been created and if not, create one.
	err := db.Init()
	if err != nil {
		log.Fatal(err)
	}
	// If you want to delete the fsentry repository you can user Drop() method.
	// db.Drop()

	/* Folder */

	// Create a folder in the root of the storage.
	_, err = db.CreateFolder("f1", nil)
	if err != nil {
		log.Fatal(err)
	}
	// Create a folder "f2" inside the folder "f1".
	_, err = db.CreateFolder("f2", Data{"folder", 17}, "f1")
	if err != nil {
		log.Fatal(err)
	}
	// Create a "f3" folder inside the "f2" folder, which is inside the "f1" folder.
	_, err = db.CreateFolder("f3", nil, "f1", "f2")
	if err != nil {
		log.Fatal(err)
	}

	/* Entry */

	// Create an entry in the root of the repository.
	err = db.CreateEntry("e1", Data{"hello", 10})
	if err != nil {
		log.Fatal(err)
	}
	// Create an entry in the "f1" folder.
	err = db.CreateEntry("e2", Data{"bye", 20}, "f1")
	if err != nil {
		log.Fatal(err)
	}

	/* Binary */

	// Create a binary file in the root of the repository.
	err = db.CreateBinary("b1", []byte("some binary data"))
	if err != nil {
		log.Fatal(err)
	}
	// Create a binary file in the "f2" folder.
	err = db.CreateBinary("b2", []byte("more binary data"), "f1", "f2")
	if err != nil {
		log.Fatal(err)
	}
}
