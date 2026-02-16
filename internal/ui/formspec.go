package ui

import (
	"github.com/charmbracelet/huh"
	"github.com/hecate-social/hecate-tui/internal/theme"
)

// FieldType identifies the kind of form field.
type FieldType int

const (
	FieldText     FieldType = iota // Single-line text input
	FieldSelect                    // Dropdown / select
	FieldTextarea                  // Multi-line text input
)

// FieldSpec declaratively describes a single form field.
type FieldSpec struct {
	Key         string
	Label       string
	Description string
	Placeholder string
	FieldType   FieldType
	Required    bool
	Default     string
	Options     []string // For FieldSelect
}

// FormSpec declaratively describes a complete form.
type FormSpec struct {
	ID     string // e.g. "devops.design_aggregate"
	Title  string
	Fields []FieldSpec
}

// BuildForm creates a FormModel from a declarative FormSpec.
// It maps each FieldSpec to the appropriate huh component and appends
// a Submit/Cancel confirmation at the end.
func BuildForm(spec FormSpec, t *theme.Theme, s *theme.Styles) *FormModel {
	// Collect value pointers so huh can bind them.
	// We store them by key in a map for extraction later.
	values := make(map[string]*string, len(spec.Fields))
	var confirm bool

	var fields []huh.Field
	for _, f := range spec.Fields {
		val := new(string)
		*val = f.Default
		values[f.Key] = val

		switch f.FieldType {
		case FieldSelect:
			opts := make([]huh.Option[string], 0, len(f.Options))
			for _, o := range f.Options {
				opts = append(opts, huh.NewOption(o, o))
			}
			fields = append(fields, huh.NewSelect[string]().
				Key(f.Key).
				Title(f.Label).
				Description(f.Description).
				Options(opts...).
				Value(val))

		case FieldTextarea:
			fields = append(fields, huh.NewText().
				Key(f.Key).
				Title(f.Label).
				Description(f.Description).
				Placeholder(f.Placeholder).
				Value(val))

		default: // FieldText
			fields = append(fields, huh.NewInput().
				Key(f.Key).
				Title(f.Label).
				Description(f.Description).
				Placeholder(f.Placeholder).
				Value(val))
		}
	}

	// Append confirm
	fields = append(fields, huh.NewConfirm().
		Key("confirm").
		Title("").
		Affirmative("Submit").
		Negative("Cancel").
		Value(&confirm))

	form := huh.NewForm(
		huh.NewGroup(fields...),
	).WithTheme(huh.ThemeCharm()).
		WithWidth(55).
		WithShowHelp(false)

	return &FormModel{
		form:     form,
		theme:    t,
		styles:   s,
		formID:   spec.ID,
		title:    spec.Title,
		width:    55,
		spec:     &spec,
		valuePtrs: values,
	}
}

// VentureInitSpec returns the FormSpec for creating a new venture.
func VentureInitSpec(cwd string) FormSpec {
	cwdDisplay := shortenHome(cwd)
	return FormSpec{
		ID:    "venture_init",
		Title: "New Venture",
		Fields: []FieldSpec{
			{
				Key:         "path",
				Label:       "Path",
				Description: "Directory to create (relative or absolute)",
				Placeholder: cwdDisplay + "/my-venture",
				FieldType:   FieldText,
			},
			{
				Key:         "name",
				Label:       "Name",
				Description: "Leave empty to use directory name",
				Placeholder: "(auto from path)",
				FieldType:   FieldText,
			},
			{
				Key:         "brief",
				Label:       "Brief",
				Description: "Optional description",
				Placeholder: "A revolutionary new product...",
				FieldType:   FieldText,
			},
		},
	}
}
