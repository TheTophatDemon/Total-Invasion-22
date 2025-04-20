#version 330

layout(location = 0) in vec3 aPos;
layout(location = 1) in vec2 aTexCoord;

uniform mat4 uViewMatrix;
uniform mat4 uProjMatrix;

out vec2 vTexCoord;

void main() {
    vTexCoord = aTexCoord;
    mat4 viewMatrixWithoutTranslation = uViewMatrix;
    viewMatrixWithoutTranslation[3] = vec4(0.0, 0.0, 0.0, 1.0);
    gl_Position = uProjMatrix * viewMatrixWithoutTranslation * vec4(aPos, 1);
}