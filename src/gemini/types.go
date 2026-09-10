package gemini

import (
	"github.com/lagosproject/StockTaggerIA/src/metadata"
)

// MetadataEntry matches the JSON structure expected by StockTaggerIA.
type MetadataEntry struct {
	Filename      string      `json:"filename"`
	Tags          interface{} `json:"tags,omitempty"`
	TagsEn        interface{} `json:"tags_en,omitempty"`
	TagsEs        interface{} `json:"tags_es,omitempty"`
	TagsFr        interface{} `json:"tags_fr,omitempty"`
	Description   interface{} `json:"description,omitempty"`
	DescriptionEn interface{} `json:"description_en,omitempty"`
	DescriptionEs interface{} `json:"description_es,omitempty"`
	DescriptionFr interface{} `json:"description_fr,omitempty"`
}

// ToMetadataRaw converts a Gemini MetadataEntry to metadata.MetadataRaw.
func (e MetadataEntry) ToMetadataRaw() metadata.MetadataRaw {
	return metadata.MetadataRaw{
		Filename:      e.Filename,
		Tags:          e.Tags,
		TagsEn:        e.TagsEn,
		TagsEs:        e.TagsEs,
		TagsFr:        e.TagsFr,
		Description:   e.Description,
		DescriptionEn: e.DescriptionEn,
		DescriptionEs: e.DescriptionEs,
		DescriptionFr: e.DescriptionFr,
	}
}

// GeminiRequest represents the payload for the Gemini generateContent endpoint.
type GeminiRequest struct {
	Contents         []Content         `json:"contents"`
	GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
}

// Content holds parts for a conversation turn.
type Content struct {
	Parts []Part `json:"parts"`
}

// Part represents a text or inline data segment.
type Part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inline_data,omitempty"`
}

// InlineData carries base64 image data and its MIME type.
type InlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

// GenerationConfig sets parameters like response MIME type.
type GenerationConfig struct {
	Temperature      float64 `json:"temperature,omitempty"`
	ResponseMimeType string  `json:"responseMimeType,omitempty"`
}

// GeminiResponse models the response from generateContent.
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}
