package render

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/failure"
)

type Framebuffer struct {
	frameBufferId uint32
	RenderTexture *textures.Texture
	depthBufferId uint32
}

func NewFramebuffer(width, height int) (buffer Framebuffer) {
	gl.GenFramebuffers(1, &buffer.frameBufferId)
	failure.CheckOpenGLError()
	gl.BindFramebuffer(gl.FRAMEBUFFER, buffer.frameBufferId)
	failure.CheckOpenGLError()

	buffer.RenderTexture = textures.GenerateRenderTexture(width, height)
	buffer.RenderTexture.Bind()

	gl.GenRenderbuffers(1, &buffer.depthBufferId)
	failure.CheckOpenGLError()
	gl.BindRenderbuffer(gl.RENDERBUFFER, buffer.depthBufferId)
	failure.CheckOpenGLError()
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT, int32(width), int32(height))
	failure.CheckOpenGLError()
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, buffer.depthBufferId)
	failure.CheckOpenGLError()
	gl.FramebufferTexture(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, buffer.RenderTexture.ID(), 0)
	failure.CheckOpenGLError()
	var drawBuffers uint32 = gl.COLOR_ATTACHMENT0
	gl.DrawBuffers(1, &drawBuffers)
	failure.CheckOpenGLError()
	if status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER); status != gl.FRAMEBUFFER_COMPLETE {
		failure.LogErrWithLocation("frame buffer is not complete: %v", status)
	}
	return
}

func (buffer Framebuffer) Bind() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, buffer.frameBufferId)
	gl.Viewport(0, 0, int32(buffer.RenderTexture.Width()), int32(buffer.RenderTexture.Height()))
	gl.ClearColor(0.0, 0.0, 0.2, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	failure.CheckOpenGLError()
}

func (buffer Framebuffer) Unbind() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	failure.CheckOpenGLError()
}

func (buffer *Framebuffer) Free() {
	if buffer.depthBufferId != 0 {
		gl.DeleteRenderbuffers(1, &buffer.depthBufferId)
	}
	if buffer.RenderTexture != nil {
		buffer.RenderTexture.Free()
	}
	if buffer.frameBufferId != 0 {
		gl.DeleteFramebuffers(1, &buffer.frameBufferId)
	}
	*buffer = Framebuffer{}
}
