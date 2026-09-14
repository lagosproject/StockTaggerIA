package metadata

import (
	"os"
	"testing"
)

func TestParseMetadataEntry(t *testing.T) {
	entry := MetadataRaw{
		Filename:      "photo1.jpg",
		TagsEn:        "mountain, landscape, nature, sunset, nature, MOUNTAIN",
		TagsEs:        "montaña, paisaje, naturaleza, atardecer",
		TagsFr:        "montagne, paysage, nature, coucher de soleil",
		DescriptionEn: "A majestic mountain at sunset.",
		DescriptionEs: "Una montaña majestuosa al atardecer.",
		DescriptionFr: "Une montagne majestueuse au coucher du soleil.",
	}

	// Test English parsing and deduplication
	tagsEn, descEn := ParseMetadataEntry(entry, "en", 30)
	if descEn != "A majestic mountain at sunset." {
		t.Errorf("Expected English description, got %s", descEn)
	}
	// "mountain" and "nature" appear twice with casing differences
	expectedTagsEn := 4 // mountain, landscape, nature, sunset
	if len(tagsEn) != expectedTagsEn {
		t.Errorf("Expected %d deduplicated tags, got %d: %v", expectedTagsEn, len(tagsEn), tagsEn)
	}

	// Test Spanish parsing
	tagsEs, descEs := ParseMetadataEntry(entry, "es", 25)
	if descEs != "Una montaña majestuosa al atardecer." {
		t.Errorf("Expected Spanish description, got %s", descEs)
	}
	if len(tagsEs) != 4 {
		t.Errorf("Expected 4 Spanish tags, got %d", len(tagsEs))
	}

	// Test French parsing
	tagsFr, descFr := ParseMetadataEntry(entry, "fr", 2)
	if descFr != "Une montagne majestueuse au coucher du soleil." {
		t.Errorf("Expected French description, got %s", descFr)
	}
	// Test maxTags limit = 2
	if len(tagsFr) != 2 {
		t.Errorf("Expected maxTags to truncate to 2 tags, got %d: %v", len(tagsFr), tagsFr)
	}
}

func TestParseStringOrSlice(t *testing.T) {
	// String with commas
	res1 := ParseStringOrSlice("one, two, three, ")
	if len(res1) != 3 || res1[0] != "one" || res1[1] != "two" || res1[2] != "three" {
		t.Errorf("Failed parsing string list: %v", res1)
	}

	// Slice of interface
	res2 := ParseStringOrSlice([]interface{}{"alpha", "beta"})
	if len(res2) != 2 || res2[0] != "alpha" || res2[1] != "beta" {
		t.Errorf("Failed parsing slice interface: %v", res2)
	}

	// Slice of strings
	res3 := ParseStringOrSlice([]string{"x", "y"})
	if len(res3) != 2 {
		t.Errorf("Failed parsing string slice: %v", res3)
	}
}

func TestCleanBaseName(t *testing.T) {
	cases := map[string]string{
		`C:\Users\Photo\Beach.jpg`: "beach.jpg",
		`/home/user/photos/Sunset.JPG`: "sunset.jpg",
		`simple.png`: "simple.png",
		`D:\Data\Sub\test.jpeg`: "test.jpeg",
	}
	for in, expected := range cases {
		out := cleanBaseName(in)
		if out != expected {
			t.Errorf("cleanBaseName(%q) = %q, expected %q", in, out, expected)
		}
	}
}

func TestLoadMetadataMapFromJSON(t *testing.T) {
	importOS := func(content string) string {
		f, err := os.CreateTemp("", "tag_test_*.json")
		if err != nil {
			t.Fatal(err)
		}
		f.WriteString(content)
		f.Close()
		return f.Name()
	}

	// 1. Array with UTF-8 BOM
	bomJSON := "\xef\xbb\xbf[{\"filename\": \"img1.jpg\", \"tags\": \"sky, blue\"}]"
	p1 := importOS(bomJSON)
	defer os.Remove(p1)

	m1, err := LoadMetadataMapFromJSON(p1)
	if err != nil {
		t.Fatalf("Failed loading BOM JSON: %v", err)
	}
	if len(m1) != 1 || m1["img1.jpg"].Filename != "img1.jpg" {
		t.Errorf("Unexpected m1 contents: %+v", m1)
	}

	// 2. Map keyed by filename
	mapJSON := "{\"img2.jpg\": {\"tags_en\": \"sea, water\"}}"
	p2 := importOS(mapJSON)
	defer os.Remove(p2)

	m2, err := LoadMetadataMapFromJSON(p2)
	if err != nil {
		t.Fatalf("Failed loading map JSON: %v", err)
	}
	if len(m2) != 1 || m2["img2.jpg"].TagsEn != "sea, water" {
		t.Errorf("Unexpected m2 contents: %+v", m2)
	}

	// 3. Wrapped object: {"images": [...]}
	wrappedJSON := "{\"images\": [{\"filename\": \"D:\\\\Photos\\\\Img3.jpg\", \"tags\": \"forest\"}]}"
	p3 := importOS(wrappedJSON)
	defer os.Remove(p3)

	m3, err := LoadMetadataMapFromJSON(p3)
	if err != nil {
		t.Fatalf("Failed loading wrapped JSON: %v", err)
	}
	if len(m3) != 1 || m3["img3.jpg"].Tags != "forest" {
		t.Errorf("Unexpected m3 contents: %+v", m3)
	}
}
