#version 330

layout(location = 0) in vec3 aPos;
layout(location = 1) in vec2 aTexCoord;
layout(location = 2) in vec3 aNormal;

uniform mat4 uViewMatrix;
uniform mat4 uProjMatrix;
uniform mat4 uModelMatrix;
uniform vec4 uSourceRect;
uniform bool uFlipHorz;

out vec2 vTexCoord;
out vec3 vNormal;

void main() {
    float sOfs = aTexCoord.x * uSourceRect.z;
    if (uFlipHorz) {
        sOfs = uSourceRect.z - sOfs;
    }
    vTexCoord = uSourceRect.xy + vec2(sOfs, -aTexCoord.y * uSourceRect.w);
    
    vec4 pos = uViewMatrix * uModelMatrix * vec4(0.0, 0.0, 0.0, 1.0);
    float scale = sqrt(uModelMatrix[0][0] * uModelMatrix[0][0] + uModelMatrix[1][0] * uModelMatrix[1][0] + uModelMatrix[2][0] * uModelMatrix[2][0]);
    pos += vec4(aPos.x * scale, aPos.y * scale, 0.0, 0.0);
    pos = uProjMatrix * pos;
    
    gl_Position = pos;
}