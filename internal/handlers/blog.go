package handlers

import (
	"net/http"
	"strings"

	"github.com/denniskbijo/visa-tracker/internal/blog"
)

func (h *Handler) handleBlog(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/blog" || r.URL.Path == "/blog/" {
		h.render(w, "blog.html", struct {
			Title     string
			ActiveNav string
			Posts     []blog.Post
		}{
			Title:     "Blog",
			ActiveNav: "blog",
			Posts:     blog.All(),
		})
		return
	}

	slug := strings.Trim(strings.TrimPrefix(r.URL.Path, "/blog/"), "/")
	post := blog.BySlug(slug)
	if post == nil {
		http.NotFound(w, r)
		return
	}

	page := "blog_" + strings.ReplaceAll(slug, "-", "_") + ".html"
	if slug == "salary-thresholds-2025" {
		page = "blog_salary_refresh.html"
	}
	if _, ok := h.tmpl[page]; !ok {
		http.NotFound(w, r)
		return
	}

	h.render(w, page, struct {
		Title     string
		ActiveNav string
		Post      blog.Post
	}{
		Title:     post.Title,
		ActiveNav: "blog",
		Post:      *post,
	})
}
