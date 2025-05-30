#version 330

layout(location = 0) in vec3 aPos;
layout(location = 1) in vec2 aTexCoord;
layout(location = 2) in vec3 aNormal;

uniform mat4 uViewMatrix;
uniform mat4 uProjMatrix;
uniform mat4 uModelMatrix;
uniform vec4 uSourceRect;

out vec2 vTexCoord;
out vec3 vNormal;

void main() {
    vTexCoord = uSourceRect.xy + vec2(aTexCoord.x * uSourceRect.z, (aTexCoord.y * uSourceRect.w) + uSourceRect.w);
    mat3 rot = mat3(uModelMatrix[0].xyz, uModelMatrix[1].xyz, uModelMatrix[2].xyz);
    vNormal = normalize(rot * aNormal);
    gl_Position = uProjMatrix * uViewMatrix * uModelMatrix * vec4(aPos, 1);
}