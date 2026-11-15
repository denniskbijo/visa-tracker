package blog

type Post struct {
	Slug    string
	Title   string
	Date    string
	Summary string
}

var posts = []Post{
	{
		Slug:    "ict-threshold-52500",
		Title:   "The ICT floor is £52,500. This site still said £48,500.",
		Date:    "2026-11-15",
		Summary: "Senior or Specialist Worker moved from £48,500 to £52,500 on 22 July 2025. The monthly gov.uk check caught the old figure, and the freshness date had gone stale.",
		// Source: internal/blog/ict-threshold-52500.md
	},
	{
		Slug:    "salary-thresholds-2025",
		Title:   "The numbers on this site were wrong. Here is the refresh.",
		Date:    "2026-09-04",
		Summary: "I launched in 2026 with April 2024 figures. Skilled Worker is now £41,700, and software developers are SOC 2134 at £54,700, not 2136.",
		// Draft source: internal/blog/salary-thresholds-2025.md
	},
}

func All() []Post {
	return posts
}

func BySlug(slug string) *Post {
	for i := range posts {
		if posts[i].Slug == slug {
			return &posts[i]
		}
	}
	return nil
}
