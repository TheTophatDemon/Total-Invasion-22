#version 330

in vec2 vTexCoord;

uniform sampler2D uTex;
uniform sampler2DArray uAtlas;
uniform int uFrame = 0;
uniform bool uAtlasUsed = false;

uniform float uFogStart;
uniform float uFogLength;

out vec4 oColor;

void main() {
    //Sample texture or atlas
    vec4 diffuse;
    if (uAtlasUsed) {
        diffuse = texture(uAtlas, vec3(vTexCoord, uFrame));
    } else {
        diffuse = texture(uTex, vTexCoord);
    }
    
    //Discard transparent pixels
    if (diffuse.a < 0.5) {
        discard;
    }
    
    //Apply depth based fog
    float depth = gl_FragCoord.z / gl_FragCoord.w;
    float fog = 1.0 - clamp((depth - uFogStart) / uFogLength, 0.0, 1.0);
    diffuse.rgb *= fog;

    oColor = diffuse;
}