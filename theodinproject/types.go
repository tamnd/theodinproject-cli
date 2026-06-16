package theodinproject

// Path is a learning path on The Odin Project.
type Path struct {
	Rank  int    `json:"rank"  csv:"rank"  tsv:"rank"`
	Slug  string `json:"slug"  csv:"slug"  tsv:"slug"`
	Title string `json:"title" csv:"title" tsv:"title"`
	URL   string `json:"url"   csv:"url"   tsv:"url"`
}

// Lesson is one lesson within a path's course.
type Lesson struct {
	Rank   int    `json:"rank"   csv:"rank"   tsv:"rank"`
	Slug   string `json:"slug"   csv:"slug"   tsv:"slug"`
	Course string `json:"course" csv:"course" tsv:"course"`
	Title  string `json:"title"  csv:"title"  tsv:"title"`
	URL    string `json:"url"    csv:"url"    tsv:"url"`
}

// Info is site-level stats.
type Info struct {
	Site   string `json:"site"   csv:"site"   tsv:"site"`
	Paths  int    `json:"paths"  csv:"paths"  tsv:"paths"`
	Source string `json:"source" csv:"source" tsv:"source"`
}
