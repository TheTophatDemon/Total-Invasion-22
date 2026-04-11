#version 330

in vec2 vTexCoord;

uniform sampler2D uTex;

layout(location = 0) out vec4 oColor;

void main() {
    oColor = texture(uTex, vTexCoord);
}