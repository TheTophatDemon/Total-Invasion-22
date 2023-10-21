#version 330

in vec2 vTexCoord;

uniform vec4 uDiffuseColor;

uniform sampler2D uTex;
uniform vec4 uSourceRect;

out vec4 oColor;

void main() {
    //Sample texture or atlas
    vec2 realTexCoord = uSourceRect.xy + (vTexCoord * uSourceRect.zw);
    vec4 diffuse = texture(uTex, realTexCoord);
    
    diffuse *= uDiffuseColor;
    if (diffuse.a <= 0.01) {
        discard;
    }

    oColor = diffuse;
}