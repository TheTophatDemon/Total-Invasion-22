-- Used to export animated sprites from Aseprite with layer & angle information
-- Currently, you must have one layer per angle

local spr = app.activeSprite
if not spr then return print('No active sprite') end

-- Show all layers so they will be exported
originalLayerVisibility = {}
for _, layer in ipairs(spr.layers) do
	originalLayerVisibility[layer.name] = layer.isVisible
	layer.isVisible = true
end

local path, title = spr.filename:match("^(.+[/\\])(.-)%.([^%.]*)$")

-- Export main sprite sheet and data
app.command.ExportSpriteSheet{
	ui=false,
	type=SpriteSheetType.PACKED,
	textureFilename=string.format("%s/%s_atlas.png", path, title),
	dataFilename=string.format("%s/%s.json", path, title),
	dataFormat=SpriteSheetDataFormat.JSON_ARRAY,
	filenameFormat="{title}_atlas.png;{layer};{frame}",
	borderPadding=1,
	mergeDuplicates=true,
	splitLayers=true,
	listLayers=true,
	listTags=true,
	listSlices=true,
}

-- Export the first frame of the front layer as the preview image
previewLayer = nil
for _, layer in ipairs(spr.layers) do
	if string.lower(layer.name) == "front" then
		previewLayer = layer
	end
	layer.isVisible = false
end
if previewLayer == nil then previewLayer = spr.layers[1] end
previewLayer.isVisible = true
previewImg = Image(spr.width, spr.height, ColorMode.RGB)
previewImg:drawSprite(spr, 1)
previewImg:saveAs(string.format("%s/%s.png", path, title))

-- Restore visibility of layers
for _, layer in ipairs(spr.layers) do
	layer.isVisible = originalLayerVisibility[layer.name] or true
end