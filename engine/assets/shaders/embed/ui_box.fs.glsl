#version 330

in vec2 vTexCoord;

uniform vec4 uDiffuseColor;

uniform sampler2D uTex;

out vec4 oColor;

void main() {
    //Sample texture
    vec4 diffuse = texture(uTex, vTexCoord);
    
    diffuse *= uDiffuseColor;
    if (diffuse.a <= 0.01) {
        discard;
    }

    oColor = diffuse;
}