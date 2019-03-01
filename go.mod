module github.com/z0rr0/ytapigo

replace github.com/z0rr0/ytapigo/ytapi => ./ytapi

replace github.com/z0rr0/ytapigo/ytapi/cloud => ./ytapi/cloud

require (
	github.com/z0rr0/ytapigo/ytapi v0.0.0-20190301092552-aedc106d24de
	github.com/z0rr0/ytapigo/ytapi/cloud v0.0.0-20190301092552-aedc106d24de // indirect
)
