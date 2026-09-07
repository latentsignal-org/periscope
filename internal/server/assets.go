package server

// safeImageTypes maps extensions to MIME types for passive
// image formats. Active content (svg, html, js) is never
// served inline to prevent stored XSS.
var safeImageTypes = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".webp": "image/webp",
	".gif":  "image/gif",
}
