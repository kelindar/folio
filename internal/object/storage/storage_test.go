package storage

import (
	"testing"

	"github.com/kelindar/ultima-web/internal/object"
	"github.com/stretchr/testify/assert"
)

func TestInsert(t *testing.T) {
	testStorage(func(db Storage, _ object.Registry) {
		app, err := object.New[*App]("my_project")
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
	testStorage(func(db Storage, _ object.Registry) {
		app, err := object.New[*App]("my_project")
		assert.NoError(t, err)

		// Insert the object
		created, err := Insert(db, app, "test")
		assert.NoError(t, err)

		// Update the object
		updated, err := Update(db, created, "test")
		assert.NoError(t, err)
		assert.Equal(t, app.URN(), updated.URN())
	})
}

func TestUpdate_Conflict(t *testing.T) {
	testStorage(func(db Storage, _ object.Registry) {
		app, err := object.New[*App]("my_project")
		assert.NoError(t, err)

		// Insert the object
		_, err = Insert(db, app, "test")
		assert.NoError(t, err)

		// Update the object
		_, err = Update(db, app, "test")
		assert.True(t, IsConflict(err))
	})
}

func TestDelete(t *testing.T) {
	testStorage(func(db Storage, _ object.Registry) {
		app, err := object.New[*App]("my_project")
		assert.NoError(t, err)

		// Insert the object
		created, err := db.Insert(app, "test")
		assert.NoError(t, err)

		// Delete the object
		deleted, err := Delete[*App](db, created.URN(), "test")
		assert.NoError(t, err)
		assert.Equal(t, app.URN(), deleted.URN())
	})
}

func TestSearch(t *testing.T) {
	testStorage(func(db Storage, _ object.Registry) {
		for i := 0; i < 10; i++ {
			app, err := object.New[*App]("my_project")
			assert.NoError(t, err)

			_, err = Insert(db, app, "test")
			assert.NoError(t, err)
		}

		results, err := Search[*App](db, Query{
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

// ---------------------------------- Storage Test ----------------------------------

func testStorage(fn func(db Storage, registry object.Registry)) {
	r := newRegistry()
	s := OpenEphemeral(r)
	defer s.Close()
	fn(s, r)
}

// ---------------------------------- Test Types ----------------------------------

type Artifact struct {
	object.Meta `kind:"artifact" json:",inline"`
	Deployment  object.URN `json:"deployment"`
}

type Deployment struct {
	object.Meta `kind:"deployment" json:",inline"`
	Env         string     `json:"env"`
	App         object.URN `json:"app"`
}

type App struct {
	object.Meta `kind:"app" json:",inline"`
}

func newRegistry() object.Registry {
	registry := object.NewRegistry()
	object.Register[*Artifact](registry)
	object.Register[*Deployment](registry)
	object.Register[*App](registry)
	return registry
}
