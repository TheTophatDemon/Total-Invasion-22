#version 330

in vec2 vTexCoord;
in vec4 vDiffuseColor;

uniform sampler2D uTex;

uniform float uFogStart;
uniform float uFogLength;

out vec4 oColor;

void main() {
    //Sample texture
    vec4 diffuse = texture(uTex, vTexCoord);
    
    //Discard transparent pixels
    if (diffuse.a < 0.5) {
        discard;
    }

    diffuse *= vDiffuseColor;
    
    //Apply depth based fog
    float depth = gl_FragCoord.z / gl_FragCoord.w;
    float fog = 1.0 - clamp((depth - uFogStart) / uFogLength, 0.0, 1.0);
    diffuse.rgb *= fog;

    oColor = diffuse;
}