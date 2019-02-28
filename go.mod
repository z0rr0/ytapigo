module github.com/z0rr0/ytapigo

replace github.com/z0rr0/ytapigo/ytapi => ./ytapi

replace github.com/z0rr0/ytapigo/ytapi/cloud => ./ytapi/cloud

require (
	github.com/z0rr0/ytapigo/ytapi v0.0.0-20190228201155-a38e2d8b4b1b
	github.com/z0rr0/ytapigo/ytapi/cloud v0.0.0-20190228201155-a38e2d8b4b1b // indirect
)
