package slack

import "encoding/json"

// DialogWrapper is the root object in a request to dialog.open
type DialogWrapper struct {
	TriggerID string `json:"trigger_id"`
	Dialog    Dialog `json:"dialog"`
}

// Dialog represents a dialog opened by dialog.open.
type Dialog struct {
	CallbackID     string        `json:"callback_id"`
	Title          string        `json:"title"`
	SubmitLabel    string        `json:"submit_label,omitempty"`
	NotifyOnCancel bool          `json:"notify_on_cancel,omitempty"`
	State          string        `json:"state,omitempty"`
	Elements       []interface{} `json:"elements"`
}

// TextArea represents a TextArea
type TextArea struct {
	Type        textAreaType `json:"type"`
	Label       string       `json:"label"`
	Name        string       `json:"name"`
	Placeholder string       `json:"placeholder,omitempty"`
	MaxLength   int          `json:"max_length,omitempty"`
	MinLength   int          `json:"min_length,omitempty"`
	Optional    bool         `json:"optional,omitempty"`
	Hint        string       `json:"hint,omitempty"`
	Subtype     string       `json:"subtype,omitempty"`
	Value       string       `json:"value,omitempty"`
}
type textAreaType string

func (textAreaType) MarshalJSON() ([]byte, error) {
	return json.Marshal("textarea")
}

// SelectElement represents a SelectElement
type SelectElement struct {
	Label           string         `json:"label"`
	Name            string         `json:"name"`
	Type            selectType     `json:"type"`
	DataSource      string         `json:"data_source,omitempty"`
	MinQueryLength  int            `json:"min_query_length,omitempty"`
	Placeholder     string         `json:"placeholder,omitempty"`
	Optional        bool           `json:"optional,omitempty"`
	Value           string         `json:"value,omitempty"`
	SelectedOptions []SelectOption `json:"selected_options,omitempty"`
	Options         []SelectOption `json:"options,omitempty"`
}
type selectType string

func (selectType) MarshalJSON() ([]byte, error) {
	return json.Marshal("select")
}

// SelectionOption represents a single option in a SelectElement
type SelectOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// SelectGroup represents a group of SelectOptions in a SelectElement.
type SelectGroup struct {
	Label   string         `json:"label"`
	Options []SelectOption `json:"options"`
}
