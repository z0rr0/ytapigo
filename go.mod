module github.com/z0rr0/ytapigo

replace github.com/z0rr0/ytapigo/ytapi => ./ytapi

replace github.com/z0rr0/ytapigo/ytapi/cloud => ./ytapi/cloud

require (
	github.com/z0rr0/ytapigo/ytapi v0.0.0-20190223214658-21a04a92e82e
	github.com/z0rr0/ytapigo/ytapi/cloud v0.0.0-20190223214658-21a04a92e82e // indirect
)
