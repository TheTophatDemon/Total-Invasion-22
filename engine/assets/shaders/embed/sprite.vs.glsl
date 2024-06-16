#version 330

layout(location = 0) in vec3 aPos;
layout(location = 1) in vec2 aTexCoord;
layout(location = 2) in vec3 aNormal;

//{{ if .Instanced }}
layout(location = 8) in vec3 aInstancePos;
layout(location = 9) in vec4 aInstanceColor;
layout(location = 10) in vec2 aInstanceSize;
layout(location = 11) in vec4 aInstanceSrc;
//{{ end }}

uniform mat4 uViewMatrix;
uniform mat4 uProjMatrix;
uniform mat4 uModelMatrix;
uniform bool uFlipHorz;

//{{ if .Instanced | not }}
uniform vec4 uSourceRect;
//{{ end }}

out vec2 vTexCoord;
out vec3 vNormal;
out vec4 vDiffuseColor;

void main() {

    vec4 srcRect;
    //{{ if .Instanced }}
    srcRect = aInstanceSrc;
    //{{ else }}
    srcRect = uSourceRect;
    //{{ end }}

    float sOfs = aTexCoord.x * srcRect.z;
    if (uFlipHorz) {
        sOfs = srcRect.z - sOfs;
    }
    vTexCoord = srcRect.xy + vec2(sOfs, -aTexCoord.y * srcRect.w);
    
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