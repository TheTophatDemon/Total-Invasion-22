package render

type TranslucentRender interface {
	DistanceFromScreen(context *Context) float32
	Render(context *Context)
}
