#version 330

in vec2 vTexCoord;
in vec3 vNormal;

uniform vec3 uLightDir;
uniform vec3 uAmbientColor;
uniform sampler2D uTex;
uniform vec4 uSourceRect;

uniform float uFogStart;
uniform float uFogLength;

out vec4 oColor;

void main() {
    //Sample texture or atlas
    vec2 realTexCoord = uSourceRect.xy + (vTexCoord * uSourceRect.zw);
    vec4 diffuse = texture(uTex, realTexCoord);
    
    //Discard transparent pixels
    if (diffuse.a < 0.5) {
        discard;
    }
    
    //Calulate diffuse lighting
    float lightFactor = (dot(-uLightDir, normalize(vNormal)) + 1.0) / 2.0;
    diffuse.rgb *= uAmbientColor + (vec3(1.0) - uAmbientColor) * lightFactor;
    
    //Apply depth based fog
    float depth = gl_FragCoord.z / gl_FragCoord.w;
    float fog = 1.0 - clamp((depth - uFogStart) / uFogLength, 0.0, 1.0);
    diffuse.rgb *= fog;

    oColor = diffuse;
}