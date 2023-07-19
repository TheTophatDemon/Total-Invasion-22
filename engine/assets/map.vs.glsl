#version 330

layout(location = 0) in vec3 aPos;
layout(location = 1) in vec2 aTexCoord;
layout(location = 2) in vec3 aNormal;

uniform mat4 uViewProjection;
uniform mat4 uModelTransform;

out vec2 vTexCoord;
out vec3 vNormal;

void main() {
    vTexCoord = aTexCoord;
    mat3 rot = mat3(uModelTransform[0].xyz, uModelTransform[1].xyz, uModelTransform[2].xyz);
    vNormal = normalize(rot * aNormal);
    gl_Position = uViewProjection * uModelTransform * vec4(aPos, 1);
}