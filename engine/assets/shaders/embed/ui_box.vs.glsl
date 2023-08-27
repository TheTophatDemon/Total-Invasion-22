#version 330

precision highp float;

layout(location = 0) in vec3 aPos;
layout(location = 1) in vec2 aTexCoord;

uniform mat4 uViewMatrix;
uniform mat4 uProjMatrix;
uniform mat4 uModelMatrix;

uniform vec4 uSrcRect;

out vec2 vTexCoord;

void main() {
    vTexCoord = uSrcRect.xy + vec2(0.0, 1.0) - (aTexCoord * uSrcRect.zw);

    gl_Position = uProjMatrix * uViewMatrix * uModelMatrix * vec4(aPos, 1.0);
}