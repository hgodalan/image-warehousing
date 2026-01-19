package models

// AIAnalysis represents the AI-generated analysis of an image or 3D object
type AIAnalysis struct {
	Type                   string              `json:"type"` // 2D or 3D
	PrimaryCategory        string              `json:"primary_category"`
	Description            string              `json:"description"`
	Objects                []string            `json:"objects"`
	Colors                 []string            `json:"colors"`
	Features               []Feature           `json:"features"`

	// 2D specific
	SceneType              string              `json:"scene_type,omitempty"`
	Mood                   string              `json:"mood,omitempty"`
	Style                  string              `json:"style,omitempty"`

	// 3D specific
	Lighting               string              `json:"lighting,omitempty"`
	ThreeDCharacteristics  string              `json:"three_d_characteristics,omitempty"`
	Symmetry               string              `json:"symmetry,omitempty"`
	Complexity             string              `json:"complexity,omitempty"`

	RawResponse            string              `json:"raw_response,omitempty"` // Full JSON from Gemini
}

// Feature represents a detected feature with confidence score
type Feature struct {
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
}
