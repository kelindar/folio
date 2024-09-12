package folio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmbed(t *testing.T) {
	r := newRegistry()

	// Create a new app
	app, err := New[*Kind3]("my_project")
	assert.NoError(t, err)

	// Create a new deployment with the associated app
	dep, err := New("my_project", func(r *Kind2) error {
		r.App = app.URN()
		return nil
	})
	assert.NoError(t, err)

	req, err := New("my_project", func(r *Kind4) error {
		r.After = Embed{Value: dep}
		return nil
	})
	assert.NoError(t, err)

	// Marshal the request
	data, err := ToJSON(req)
	assert.NoError(t, err)

	// Unmarshal the request
	out, err := FromJSON(r, data)
	assert.NoError(t, err)
	decoded := out.(*Kind4).After.Value.(*Kind2)
	assert.NotNil(t, decoded)
	assert.Equal(t, dep.ID, decoded.ID)

}
