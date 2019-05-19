package main

type MozThingLink struct {
	Rel       string `json:"rel"`
	Href      string `json:"href"`
	MediaType string `json:"mediaType,omitempty"`
}

type MozThingProperty struct {
	Title       string         `json:"title"`
	Type        string         `json:"type"`
	Unit        string         `json:"unit"`
	ReadOnly    bool           `json:"readOnly"`
	Description string         `json:"description"`
	Links       []MozThingLink `json:"links"`
}

type MozThingAction struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type MozThingEvent struct {
	Description string `json:"description"`
}

type MozThingDefinition struct {
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	Properties  map[string]MozThingProperty `json:"properties"`
	Actions     map[string]MozThingAction   `json:"actions"`
	Events      map[string]MozThingEvent    `json:"events"`
	Links       []MozThingLink              `json:"links"`
}
