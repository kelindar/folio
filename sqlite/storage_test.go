package sqlite

import (
	"fmt"
	"testing"

	"github.com/kelindar/folio"
	"github.com/stretchr/testify/assert"
)

func TestInsert(t *testing.T) {
	testStorage(func(db folio.Storage, _ folio.Registry) {
		app, err := folio.New[*App]("my_project")
		assert.NoError(t, err)

		// Insert the object
		out, err := db.Insert(app, "test")
		assert.NoError(t, err)
		assert.Equal(t, app.URN(), out.URN())

		createdBy, createdAt := out.Created()
		assert.NotEmpty(t, createdAt)
		assert.Equal(t, "test", createdBy)
	})
}

func TestUpdate(t *testing.T) {
	testStorage(func(db folio.Storage, _ folio.Registry) {
		app, err := folio.New[*App]("my_project")
		assert.NoError(t, err)

		// Insert the object
		created, err := db.Insert(app, "test")
		assert.NoError(t, err)

		// Update the object
		updated, err := db.Update(created, "test")
		assert.NoError(t, err)
		assert.Equal(t, app.URN(), updated.URN())
	})
}

func TestUpdate_Conflict(t *testing.T) {
	testStorage(func(db folio.Storage, _ folio.Registry) {
		app, err := folio.New[*App]("my_project")
		assert.NoError(t, err)

		// Insert the object
		_, err = db.Insert(app, "test")
		assert.NoError(t, err)

		// Reset the updated at time
		app.UpdatedAt = 0

		// Update the object
		_, err = db.Update(app, "test")
		assert.True(t, folio.IsConflict(err))
	})
}

func TestDelete(t *testing.T) {
	testStorage(func(db folio.Storage, _ folio.Registry) {
		app, err := folio.New[*App]("my_project")
		assert.NoError(t, err)

		// Insert the object
		created, err := db.Insert(app, "test")
		assert.NoError(t, err)

		// Delete the object
		deleted, err := db.Delete(created.URN(), "test")
		assert.NoError(t, err)
		assert.Equal(t, app.URN(), deleted.URN())
	})
}

func TestSearch(t *testing.T) {
	testStorage(func(db folio.Storage, _ folio.Registry) {
		for i := 0; i < 10; i++ {
			v, err := folio.New[*App]("my_project")
			assert.NoError(t, err)

			_, err = db.Insert(v, "test")
			assert.NoError(t, err)
		}

		results, err := db.Search("App", folio.Query{
			Namespace: "my_project",
			Offset:    1,
			Limit:     5,
		})
		assert.NoError(t, err)

		count := 0
		for range results {
			count++
		}
		assert.Equal(t, 5, count)
	})
}

func TestSearch_FullText(t *testing.T) {
	testStorage(func(db folio.Storage, _ folio.Registry) {
		for i := 0; i < 100; i++ {
			v, _ := folio.New[*App]("my_project")
			v.Name = fmt.Sprintf("Application number %d", i)
			_, err := db.Insert(v, "test")
			assert.NoError(t, err)
		}

		results, err := db.Search("App", folio.Query{
			Namespace: "my_project",
			Match:     "appli 47",
			Limit:     1,
		})
		assert.NoError(t, err)

		count := 0
		for result := range results {
			count++
			assert.Equal(t, "Application number 47", result.(*App).Name)
		}

		assert.Equal(t, 1, count)
	})
}

func TestCount(t *testing.T) {
	testStorage(func(db folio.Storage, _ folio.Registry) {
		for i := 0; i < 10; i++ {
			v, err := folio.New[*App]("my_project")
			assert.NoError(t, err)

			_, err = db.Insert(v, "test")
			assert.NoError(t, err)
		}

		ct, err := db.Count("App", folio.Query{
			Namespace: "my_project",
		})
		assert.NoError(t, err)
		assert.Equal(t, 10, ct)
	})
}

// ---------------------------------- Storage Test ----------------------------------

func testStorage(fn func(db folio.Storage, registry folio.Registry)) {
	r := newRegistry()
	s := OpenEphemeral(r)
	defer s.Close()
	fn(s, r)
}

// ---------------------------------- Test Types ----------------------------------

type Artifact struct {
	folio.Meta `kind:"artifact" json:",inline"`
	Deployment folio.URN `json:"deployment"`
}

type Deployment struct {
	folio.Meta `kind:"deployment" json:",inline"`
	Env        string    `json:"env"`
	App        folio.URN `json:"app"`
}

type App struct {
	folio.Meta `kind:"app" json:",inline"`
	Name       string `json:"name"`
}

func newRegistry() folio.Registry {
	registry := folio.NewRegistry()
	folio.Register[*Artifact](registry)
	folio.Register[*Deployment](registry)
	folio.Register[*App](registry)
	return registry
}
