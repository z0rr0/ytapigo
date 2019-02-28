module github.com/z0rr0/ytapigo

replace github.com/z0rr0/ytapigo/ytapi => ./ytapi

replace github.com/z0rr0/ytapigo/ytapi/cloud => ./ytapi/cloud

require (
	github.com/z0rr0/ytapigo/ytapi v0.0.0-20190227150820-43a69cb6c8bd
	github.com/z0rr0/ytapigo/ytapi/cloud v0.0.0-20190227150820-43a69cb6c8bd // indirect
)
