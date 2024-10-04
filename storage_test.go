package folio_test

import (
	"testing"

	"github.com/kelindar/folio"
	"github.com/kelindar/folio/sqlite"
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
		created, err := folio.Insert(db, app, "test")
		assert.NoError(t, err)

		// Update the object
		updated, err := folio.Update(db, created, "test")
		assert.NoError(t, err)
		assert.Equal(t, app.URN(), updated.URN())
	})
}

func TestUpdate_Conflict(t *testing.T) {
	testStorage(func(db folio.Storage, _ folio.Registry) {
		app, err := folio.New[*App]("my_project")
		assert.NoError(t, err)

		// Insert the object
		_, err = folio.Insert(db, app, "test")
		assert.NoError(t, err)
		app.UpdatedAt = 0

		// Update the object
		_, err = folio.Update(db, app, "test")
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
		deleted, err := folio.Delete[*App](db, created.URN(), "test")
		assert.NoError(t, err)
		assert.Equal(t, app.URN(), deleted.URN())
	})
}

func TestSearch(t *testing.T) {
	testStorage(func(db folio.Storage, _ folio.Registry) {
		for i := 0; i < 10; i++ {
			app, err := folio.New[*App]("my_project")
			assert.NoError(t, err)

			_, err = folio.Insert(db, app, "test")
			assert.NoError(t, err)
		}

		results, err := folio.Search[*App](db, folio.Query{
			Namespaces: []string{"my_project"},
			Limit:      5,
		})
		assert.NoError(t, err)

		count := 0
		for range results {
			count++
		}
		assert.Equal(t, 5, count)
	})
}

func TestCount(t *testing.T) {
	testStorage(func(db folio.Storage, _ folio.Registry) {
		for i := 0; i < 10; i++ {
			app, err := folio.New[*App]("my_project")
			assert.NoError(t, err)

			_, err = folio.Insert(db, app, "test")
			assert.NoError(t, err)
		}

		count, err := folio.Count[*App](db, folio.Query{
			Namespaces: []string{"my_project"},
		})
		assert.NoError(t, err)
		assert.Equal(t, 10, count)
	})
}

// ---------------------------------- Storage Test ----------------------------------

func testStorage(fn func(db folio.Storage, registry folio.Registry)) {
	r := newRegistry()
	s := sqlite.OpenEphemeral(r)
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
}

func newRegistry() folio.Registry {
	registry := folio.NewRegistry()
	folio.Register[*Artifact](registry)
	folio.Register[*Deployment](registry)
	folio.Register[*App](registry)
	return registry
}
