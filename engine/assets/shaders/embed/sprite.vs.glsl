#version 330

layout(location = 0) in vec3 aPos;
layout(location = 1) in vec2 aTexCoord;
layout(location = 2) in vec3 aNormal;

//{{ if .Instanced }}
layout(location = 8) in vec3 aInstancePos;
layout(location = 9) in vec4 aInstanceColor;
layout(location = 10) in vec2 aInstanceSize;
//{{ end }}

uniform mat4 uViewMatrix;
uniform mat4 uProjMatrix;
uniform mat4 uModelMatrix;
uniform vec4 uSourceRect;
uniform bool uFlipHorz;

out vec2 vTexCoord;
out vec3 vNormal;
out vec4 vDiffuseColor;

void main() {
    float sOfs = aTexCoord.x * uSourceRect.z;
    if (uFlipHorz) {
        sOfs = uSourceRect.z - sOfs;
    }
    vTexCoord = uSourceRect.xy + vec2(sOfs, -aTexCoord.y * uSourceRect.w);
    
    mat4 modelMatrix = uModelMatrix;

    //{{ if .Instanced }}
    modelMatrix[3] += vec4(aInstancePos, 0.0);
    //{{ end }}

    vec4 pos = uViewMatrix * modelMatrix * vec4(0.0, 0.0, 0.0, 1.0);
    float modelScale = sqrt(uModelMatrix[0][0] * uModelMatrix[0][0] + uModelMatrix[1][0] * uModelMatrix[1][0] + uModelMatrix[2][0] * uModelMatrix[2][0]);
    vec2 spriteScale = vec2(modelScale, modelScale);

    //{{ if .Instanced }}
    spriteScale *= aInstanceSize;
    //{{ end }}
    
    pos += vec4(aPos.x * spriteScale.x, aPos.y * spriteScale.y, 0.0, 0.0);
    pos = uProjMatrix * pos;
    
    //{{ if .Instanced }}
    vDiffuseColor = aInstanceColor;
    //{{ else }}
    vDiffuseColor = vec4(1.0, 1.0, 1.0, 1.0);
    //{{ end }}

    gl_Position = pos;
}