#version 330

in vec2 vTexCoord;

uniform vec4 uDiffuseColor;

uniform sampler2D uTex;
uniform sampler2DArray uAtlas;
uniform int uFrame = 0;
uniform bool uAtlasUsed = false;

out vec4 oColor;

void main() {
    //Sample texture or atlas
    vec4 diffuse;
    if (uAtlasUsed) {
        diffuse = texture(uAtlas, vec3(vTexCoord, uFrame));
    } else {
        diffuse = texture(uTex, vTexCoord);
    }
    
    diffuse *= uDiffuseColor;
    if (diffuse.a <= 0.01) {
        discard;
    }

    oColor = diffuse;
}