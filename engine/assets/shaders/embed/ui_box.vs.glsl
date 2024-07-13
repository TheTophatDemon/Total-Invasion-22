#version 330

precision highp float;

layout(location = 0) in vec3 aPos;
layout(location = 1) in vec2 aTexCoord;
layout(location = 3) in vec4 aColor;

uniform mat4 uViewMatrix;
uniform mat4 uProjMatrix;
uniform mat4 uModelMatrix;

uniform vec4 uSourceRect;
uniform bool uFlipHorz;

out vec2 vTexCoord;
out vec4 vColor;

void main() {
    float sOfs = aTexCoord.x * uSourceRect.z;
    if (uFlipHorz) {
        sOfs = uSourceRect.z - sOfs;
    }
    vTexCoord = uSourceRect.xy + vec2(sOfs, aTexCoord.y * uSourceRect.w - uSourceRect.w);
    vColor = aColor;

    gl_Position = uProjMatrix * uViewMatrix * uModelMatrix * vec4(aPos, 1.0);
}