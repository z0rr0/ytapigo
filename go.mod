module github.com/z0rr0/ytapigo

replace github.com/z0rr0/ytapigo/ytapi => ./ytapi

replace github.com/z0rr0/ytapigo/ytapi/cloud => ./ytapi/cloud

require (
	github.com/z0rr0/ytapigo/ytapi v0.0.0-20190228201650-0eb5a198698f
	github.com/z0rr0/ytapigo/ytapi/cloud v0.0.0-20190228201650-0eb5a198698f // indirect
)
