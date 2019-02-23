module github.com/z0rr0/ytapigo

replace github.com/z0rr0/ytapigo/ytapi => ./ytapi

replace github.com/z0rr0/ytapigo/ytapi/cloud => ./ytapi/cloud

require (
	github.com/z0rr0/ytapigo/ytapi v0.0.0-20190223215850-f58a17e60e62
	github.com/z0rr0/ytapigo/ytapi/cloud v0.0.0-20190223215850-f58a17e60e62 // indirect
)
