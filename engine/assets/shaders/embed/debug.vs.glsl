#version 330

layout(location = 0) in vec3 aPos;
layout(location = 3) in vec4 aColor;

uniform mat4 uViewMatrix;
uniform mat4 uProjMatrix;
uniform mat4 uModelMatrix;

out vec4 vColor;

void main() {
    vColor = aColor;
    gl_Position = uProjMatrix * uViewMatrix * uModelMatrix * vec4(aPos, 1);
}