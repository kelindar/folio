package render

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kelindar/folio"
	"github.com/kelindar/folio/internal/convert"
	"github.com/kelindar/folio/sqlite"
	"github.com/stretchr/testify/assert"
)

type Product struct {
	folio.Meta `kind:"product" json:",inline"`
	Name       string
}

func TestEnum(t *testing.T) {
	test := &struct {
		Field string `is:"in(one|two|three)"`
	}{
		Field: "one",
	}

	rv := reflect.Indirect(reflect.ValueOf(test))
	field, _ := rv.Type().FieldByName("Field")
	fv := rv.FieldByName("Field")

	enum := lookupForEnum(fv)
	assert.True(t, enum.Init(&Props{
		Field: field,
	}))
	assert.NotNil(t, enum)

	k, v := enum.Current()
	assert.Equal(t, "one", k)
	assert.Equal(t, "One", v)
	assert.Equal(t, []string{"one", "two", "three"}, enum.choices)

	var keys, values []string
	for k, v := range enum.Choices() {
		keys = append(keys, k)
		values = append(values, v)
	}

	assert.Equal(t, []string{"one", "two", "three"}, keys)
	assert.Equal(t, []string{"One", "Two", "Three"}, values)

}

func TestLongEnum(t *testing.T) {
	test := &struct {
		Field string `is:"required,in(one_handed|two_handed|shoes|pants|shirt|helm|gloves|ring|talisman|neck|hair|waist|inner_torso|bracelet|face|facial_hair|middle_torso|earrings|arms|cloak|backpack|outer_torso|outer_legs|inner_legs|mount|shop_buy|shop_resale|shop_sell|bank|secure_trade)"`
	}{
		Field: "one_handed",
	}

	rv := reflect.Indirect(reflect.ValueOf(test))
	field, _ := rv.Type().FieldByName("Field")
	fv := rv.FieldByName("Field")

	enum := lookupForEnum(fv)
	assert.True(t, enum.Init(&Props{
		Field: field,
	}))
	assert.NotNil(t, enum)

	var keys, values []string
	for k, v := range enum.Choices() {
		keys = append(keys, k)
		values = append(values, v)
	}

	assert.Equal(t, 30, len(keys))
}

func TestUrn(t *testing.T) {
	testStorage(func(db folio.Storage, registry folio.Registry) {
		phone, err := folio.New("default", func(obj *Product) error {
			obj.Name = "phone"
			return nil
		})
		assert.NoError(t, err)
		assert.NotNil(t, phone)

		_, err = db.Insert(phone, "test")
		assert.NoError(t, err)
		assert.NotNil(t, phone)

		test := &struct {
			folio.Meta `kind:"mock" json:",inline"`
			Product    folio.URN `is:"required" kind:"product"`
		}{Product: phone.URN()}

		// Find the product field
		rv := reflect.Indirect(reflect.ValueOf(test))
		field, _ := rv.Type().FieldByName("Product")
		fv := rv.FieldByName("Product") // Get the field value

		// Create a lookup for the urn & init
		enum := lookupForUrn()
		assert.True(t, enum.Init(&Props{
			Context: &Context{
				Registry: registry,
				Store:    db,
			},
			Value:  fv, // Use the field value, not the struct
			Field:  field,
			Parent: test,
		}))
		assert.NotNil(t, enum)
	})
}

func TestUrnSlice(t *testing.T) {
	testStorage(func(db folio.Storage, registry folio.Registry) {
		// Create some test products
		phone, err := folio.New("default", func(obj *Product) error {
			obj.Name = "phone"
			return nil
		})
		assert.NoError(t, err)
		assert.NotNil(t, phone)

		laptop, err := folio.New("default", func(obj *Product) error {
			obj.Name = "laptop"
			return nil
		})
		assert.NoError(t, err)
		assert.NotNil(t, laptop)

		tablet, err := folio.New("default", func(obj *Product) error {
			obj.Name = "tablet"
			return nil
		})
		assert.NoError(t, err)
		assert.NotNil(t, tablet)

		// Insert products into the database
		_, err = db.Insert(phone, "test")
		assert.NoError(t, err)
		_, err = db.Insert(laptop, "test")
		assert.NoError(t, err)
		_, err = db.Insert(tablet, "test")
		assert.NoError(t, err)

		// Create a test struct with a slice of URNs
		test := &struct {
			folio.Meta `kind:"mock" json:",inline"`
			Inventory  []folio.URN `kind:"product"`
		}{
			Inventory: []folio.URN{phone.URN(), laptop.URN()},
		}

		// Find the inventory field
		rv := reflect.Indirect(reflect.ValueOf(test))
		field, _ := rv.Type().FieldByName("Inventory")
		fv := rv.FieldByName("Inventory")

		// Create a lookup for the URN slice & initialize it
		lookup := lookupForUrnSlice()
		props := &Props{
			Context: &Context{
				Registry:  registry,
				Store:     db,
				Namespace: "default",
			},
			Value:  fv,
			Field:  field,
			Parent: test,
		}

		// Test initialization
		assert.True(t, lookup.Init(props))
		assert.NotNil(t, lookup)
		assert.Equal(t, 2, len(lookup.objects))

		// Test Current - should return empty values since multiple are selected
		k, v := lookup.Current()
		assert.Equal(t, "", k)
		assert.Equal(t, "", v)

		// Test Choices - should return all available products
		var keys, values []string
		for k, v := range lookup.Choices() {
			keys = append(keys, k)
			values = append(values, v)
		}
		assert.Equal(t, 3, lookup.Len())
		assert.Equal(t, 3, len(keys))

		// Verify the objects were loaded correctly
		assert.Equal(t, "phone", lookup.objects[0].(*Product).Name)
		assert.Equal(t, "laptop", lookup.objects[1].(*Product).Name)

		// Test initialization with a non-URN slice
		invalidTest := &struct {
			folio.Meta `kind:"mock" json:",inline"`
			Tags       []string
		}{
			Tags: []string{"tag1", "tag2"},
		}

		rv = reflect.Indirect(reflect.ValueOf(invalidTest))
		field, _ = rv.Type().FieldByName("Tags")
		fv = rv.FieldByName("Tags")

		props.Value = fv
		props.Field = field
		props.Parent = invalidTest

		// Should not initialize for non-URN slice
		assert.False(t, lookup.Init(props))
	})
}

func TestUrnSliceEmpty(t *testing.T) {
	testStorage(func(db folio.Storage, registry folio.Registry) {
		// Create a test struct with an empty slice of URNs
		test := &struct {
			folio.Meta `kind:"mock" json:",inline"`
			Inventory  []folio.URN `kind:"product"`
		}{
			Inventory: []folio.URN{},
		}

		// Find the inventory field
		rv := reflect.Indirect(reflect.ValueOf(test))
		field, _ := rv.Type().FieldByName("Inventory")
		fv := rv.FieldByName("Inventory")

		// Create a lookup for the URN slice & initialize it
		lookup := lookupForUrnSlice()
		props := &Props{
			Context: &Context{
				Registry:  registry,
				Store:     db,
				Namespace: "default",
			},
			Value:  fv,
			Field:  field,
			Parent: test,
		}

		// Test initialization with empty slice
		assert.True(t, lookup.Init(props))
		assert.NotNil(t, lookup)
		assert.Equal(t, 0, len(lookup.objects))

		// Test with no kind tag
		invalidTest := &struct {
			folio.Meta `kind:"mock" json:",inline"`
			References []folio.URN
		}{
			References: []folio.URN{},
		}

		rv = reflect.Indirect(reflect.ValueOf(invalidTest))
		field, _ = rv.Type().FieldByName("References")
		fv = rv.FieldByName("References")

		props.Value = fv
		props.Field = field
		props.Parent = invalidTest

		// Should not initialize without a kind tag
		assert.False(t, lookup.Init(props))
	})
}

func TestDecodeFlags(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		expected []string
		hasFlags bool
	}{
		{"empty tag", "", nil, false},
		{"no flags", "required", nil, false},
		{"simple flags", "flags(electric|petrol|diesel)", []string{"electric", "petrol", "diesel"}, true},
		{"flags with required", "required,flags(red|blue|green)", []string{"red", "blue", "green"}, true},
		{"complex flags", "required,flags(gps|bluetooth|sunroof|leather|heated_seats),minlen(1)", []string{"gps", "bluetooth", "sunroof", "leather", "heated_seats"}, true},
		{"single flag", "flags(electric)", []string{"electric"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := reflect.StructField{
				Tag: reflect.StructTag(`is:"` + tt.tag + `"`),
			}
			result, hasFlags := decodeFlags(field)
			assert.Equal(t, tt.hasFlags, hasFlags)
			assert.EqualValues(t, tt.expected, result)
		})
	}
}

func TestRenderValueWithFlags(t *testing.T) {
	// Test structs for comprehensive flags functionality
	type TestVehicle struct {
		folio.Meta `kind:"vehicle" json:",inline"`
		Type       []string `json:"type" form:"rw" is:"flags(electric|petrol|diesel)"`
		Features   []string `json:"features" form:"rw" is:"flags(gps|bluetooth|sunroof)"`
		Colors     []string `json:"colors" form:"rw" is:"required,flags(red|blue|green)"`
		Empty      []string `json:"empty" form:"rw" is:"flags(a|b|c)"` // Empty flags field
		Regular    []string `json:"regular" form:"rw"`                 // Regular string slice without flags
	}

	// Create test vehicles with different flag states
	vehicles := map[string]*TestVehicle{
		"populated": {
			Type:     []string{"electric", "diesel"},
			Features: []string{"gps", "bluetooth"},
			Colors:   []string{"red"},
			Empty:    []string{},
			Regular:  []string{"item1", "item2"},
		},
		"single_values": {
			Type:     []string{"petrol"},
			Features: []string{"sunroof"},
			Colors:   []string{"blue"},
			Empty:    []string{},
			Regular:  []string{"single"},
		},
		"all_values": {
			Type:     []string{"electric", "petrol", "diesel"},
			Features: []string{"gps", "bluetooth", "sunroof"},
			Colors:   []string{"red", "blue", "green"},
			Empty:    []string{"a", "b", "c"},
			Regular:  []string{"all1", "all2", "all3"},
		},
	}

	// Set up registry once
	registry := folio.NewRegistry()
	typ, err := folio.Register[*TestVehicle](registry)
	assert.NoError(t, err)

	ctx := &Context{Mode: ModeEdit, Type: typ, Registry: registry}

	for testName, vehicle := range vehicles {
		t.Run(testName, func(t *testing.T) {
			rv := reflect.ValueOf(vehicle).Elem()
			rt := rv.Type()

			// Test each flags field comprehensively
			flagsFields := []string{"Type", "Features", "Colors", "Empty"}
			for _, fieldName := range flagsFields {
				field, _ := rt.FieldByName(fieldName)
				props := &Props{
					Context: ctx,
					Name:    folio.Path(strings.ToLower(fieldName)),
					Value:   rv.FieldByName(fieldName),
					Field:   field,
				}

				// Test flags lookup creation and initialization
				lookup := lookupForFlags(props.Value)
				assert.True(t, lookup.Init(props), "Flags lookup should initialize for %s", fieldName)

				// Test choices parsing
				expectedChoices := map[string][]string{
					"Type":     {"electric", "petrol", "diesel"},
					"Features": {"gps", "bluetooth", "sunroof"},
					"Colors":   {"red", "blue", "green"},
					"Empty":    {"a", "b", "c"},
				}
				assert.Equal(t, len(expectedChoices[fieldName]), lookup.Len(), "Wrong choice count for %s", fieldName)

				// Test Contains method for current values
				currentValues := props.Value.Interface().([]string)
				for _, value := range currentValues {
					assert.True(t, lookup.Contains(value), "Should contain %s in %s", value, fieldName)
				}

				// Test Current method
				currentKey, currentValue := lookup.Current()
				if len(currentValues) == 0 {
					assert.Equal(t, "", currentKey, "Empty field should have empty key")
					assert.Equal(t, "None selected", currentValue, "Empty field should show 'None selected'")
				} else {
					assert.Equal(t, strings.Join(currentValues, ","), currentKey, "Current key mismatch for %s", fieldName)
					assert.Contains(t, currentValue, convert.TitleCase(currentValues[0]), "Current value should contain first item")
				}

				// Test renderValue returns Flags component
				_, component := renderValue(props)
				assert.NotNil(t, component, "Flags component should not be nil for %s", fieldName)
			}

			// Test regular field (should NOT use flags)
			regularField, _ := rt.FieldByName("Regular")
			regularProps := &Props{
				Context: ctx,
				Name:    folio.Path("regular"),
				Value:   rv.FieldByName("Regular"),
				Field:   regularField,
			}

			// Regular field should not initialize as flags
			regularLookup := lookupForFlags(regularProps.Value)
			assert.False(t, regularLookup.Init(regularProps), "Regular field should not initialize as flags")

			// Regular field should use Strings component
			_, regularComponent := renderValue(regularProps)
			assert.NotNil(t, regularComponent, "Regular component should not be nil")
		})
	}
}

func testStorage(fn func(db folio.Storage, registry folio.Registry)) {
	r := folio.NewRegistry()
	folio.Register[*Product](r)

	s := sqlite.OpenEphemeral(r)
	defer s.Close()
	fn(s, r)
}
