#version 330

in vec2 vTexCoord;

uniform vec4 uDiffuseColor;
uniform bool uNoTexture;
uniform sampler2D uTex;

out vec4 oColor;

void main() {
    //Sample texture
    vec4 diffuse;
    
    if (!uNoTexture) {
        diffuse = texture(uTex, vTexCoord) * uDiffuseColor;
    } else {
        diffuse = uDiffuseColor;
    }
    
    if (diffuse.a <= 0.01) {
        discard;
    }

    oColor = diffuse;
}