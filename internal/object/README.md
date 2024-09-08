# Resource-Based Storage

This library provides a simple interface for storing and retrieving resources based on a document model.

## Example

First we define some resources and their relationships.

```go
type Artifact struct {
	Resource   `kind:"artifact" json:",inline"`
	Deployment URN `json:"deployment"`
}

type Deployment struct {
	Resource `kind:"deployment" json:",inline"`
	Env      string `json:"env"`
	App      URN    `json:"app"`
}

type App struct {
	Resource `kind:"app" json:",inline"`
}

type DeployReq struct {
	Resource `kind:"deploy" json:",inline"`
	Before   Embed `json:"before,omitempty"`
	After    Embed `json:"after,omitempty"`
}
```

Then we create a registry and register the resources.

```go
registry := object.NewRegistry()
object.Register[*Artifact](registry)
object.Register[*Deployment](registry)
object.Register[*App](registry)
object.Register[*DeployReq](registry)
```

Finally, we can now create those resources.

```go
app, err := record.New[*App]("organization1", "app1")
```

## Storage

The storage sub-library provides a simple interface for storing and retrieving resources based on a document model. The storage library has both generic and non-generic procedures.

```go
db, err := storage.Open(...)
defer db.Close()
// ...

app, err := record.New[*App]("organization1", "app1")
// ...

// Insert the resource into the database
inserted, err := storage.Insert(db, app, "roman")

// Fetch the resource from the database using its URN
fetched, err := storage.Fetch[*App](db, app.URN())
```
