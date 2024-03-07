# fsentry

Allows storing hierarchical data in files and folders on the file system with json descriptions and creation/update timestamps.

---
### Examples:

How to init a repository:
```go
// Initializing the fsentry repository.
db := fsentry.NewFSEntry("test",
    fsentry.WithPretty(), // If you want to store json metadata files in a pretty format.
)
// Check if a repository folder has been created and if not, create one.
err := db.Init()
if err != nil {
    t.Fatal(err)
}
// If you want to delete the fsentry repository you can user Drop() method.
// db.Drop()
```

Create a folder:
```go


// Create a folder in the root of the storage.
_, err = db.CreateFolder("f1", nil)
if err != nil {
    t.Fatal(err)
}
// Create a folder "f2" inside the folder "f1".
_, err = db.CreateFolder("f2", nil, "f1")
if err != nil {
    t.Fatal(err)
}
// Create a "f3" folder inside the "f2" folder, which is inside the "f1" folder.
_, err = db.CreateFolder("f3", nil, "f1", "f2")
if err != nil {
    t.Fatal(err)
}
```

Create an entry:
```go
// Some type with useful information that we will store inside the entry.
type Data struct {
    Title string
    Value int
}
// Create an entry in the root of the repository.
err = db.CreateEntry("e1", Data{"hello", 10})
if err != nil {
    t.Fatal(err)
}
// Create an entry in the "f1" folder.
err = db.CreateEntry("e2", Data{"bye", 20}, "f1")
if err != nil {
    t.Fatal(err)
}
```

Create a binary file:
```go
// Create a binary file in the root of the repository.
err = db.CreateBinary("b1", []byte("some binary data"))
if err != nil {
    t.Fatal(err)
}
// Create a binary file in the "f2" folder.
err = db.CreateBinary("b2", []byte("more binary data"), "f1", "f2")
if err != nil {
    t.Fatal(err)
}
```