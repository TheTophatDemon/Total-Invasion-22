package screens

import (
	"log"
	"strings"
	"text/scanner"
	"unicode/utf8"
	"unsafe"

	"github.com/TotallyGamerJet/clay"
	"github.com/fzipp/bmfont"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/containers/maybe"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type ClayTestScreen struct {
	HeeHeeCount int
	SliderValue float32
}

func (scr *ClayTestScreen) Enter() {
	arena := clay.CreateArenaWithCapacityAndMemory(make([]byte, clay.MinMemorySize()))
	width, height := engine.ScreenSize()
	clay.Initialize(arena, clay.Dimensions{Width: float32(width), Height: float32(height)}, clay.ErrorHandler{
		ErrorHandlerFunction: func(errorText clay.ErrorData) {
			log.Println(errorText.Error())
		},
	})
	clay.SetMeasureTextFunction(
		func(text clay.StringSlice, config *clay.TextElementConfig, userData unsafe.Pointer) clay.Dimensions {
			textStr := text.String()
			font := cache.DefaultFont
			var textWidth float32
			var prevRune rune
			var maxCharHeight float32
			runeIndex := 0
			for i, runeWidth := 0, 0; i < len(textStr); i += runeWidth {
				var rn rune
				rn, runeWidth = utf8.DecodeRuneInString(textStr[i:])
				char, ok := font.Chars[rn]

				if !ok {
					// Add blank space for unknown character
					textWidth += 16
					continue
				}

				// Find character position
				textWidth += float32(char.Size().X) + float32(char.XAdvance)
				maxCharHeight = max(maxCharHeight, float32(char.Size().Y))

				// Add kerning
				if prevRune != scanner.EOF {
					pair := bmfont.CharPair{First: prevRune, Second: rn}
					kerning, ok := font.Kerning[pair]
					if ok {
						textWidth += float32(kerning.Amount)
					}
				}

				prevRune = rn
				runeIndex += 1
			}
			return clay.Dimensions{
				Width:  textWidth,
				Height: maxCharHeight,
			}
		},
		unsafe.Pointer(&struct{}{}),
	)
}
func (scr *ClayTestScreen) Exit() {}

func scrollBarOnHover(elementId clay.ElementId, pointerInfo clay.PointerData, userData unsafe.Pointer) {
	scr := (*ClayTestScreen)(userData)
	if pointerInfo.State == clay.POINTER_DATA_PRESSED {
		elemData := clay.GetElementData(elementId)
		scr.SliderValue = (pointerInfo.Position.X - elemData.BoundingBox.X) / elemData.BoundingBox.Width
		println(scr.SliderValue)
	}
}

func (scr *ClayTestScreen) Layout(queue *ui.RenderQueue, deltaTime float32) {
	mousePos := input.MousePosition()
	clay.SetPointerState(clay.Vector2{
		X: mousePos[0],
		Y: mousePos[1],
	}, input.IsMouseButtonDown(glfw.MouseButton1))

	clay.UpdateScrollContainers(false, clay.Vector2{Y: input.MouseScrollDelta()}, deltaTime)

	screenWidth, screenHeight := engine.ScreenSize()
	clay.SetLayoutDimensions(clay.Dimensions{
		Width:  float32(screenWidth),
		Height: float32(screenHeight),
	})

	if settings.Current.ActionWeaponWheel.JustPressed() {
		clay.SetDebugModeEnabled(!clay.IsDebugModeEnabled())
	}

	clay.BeginLayout()

	clay.UI(clay.ID("HeaderBar"))(clay.ElementDeclaration{
		Layout: clay.LayoutConfig{
			Sizing: clay.Sizing{
				Height: clay.SizingFixed(60),
				Width:  clay.SizingGrow(0),
			},
			Padding:  clay.Padding{Left: 16, Right: 16},
			ChildGap: 16,
			ChildAlignment: clay.ChildAlignment{
				X: clay.ALIGN_X_CENTER,
			},
			LayoutDirection: clay.TOP_TO_BOTTOM,
		},
		Clip: clay.ClipElementConfig{
			Vertical:    true,
			ChildOffset: clay.GetScrollOffset(),
		},
		BackgroundColor: func() clay.Color {
			if clay.Hovered() {
				return clay.Color{R: 1.0, G: 1.0, B: 1.0, A: 1.0}
			}
			return clay.Color{R: 0.1, G: 0.9, B: 0.9, A: 1.0}
		}(),
		CornerRadius: clay.CornerRadiusAll(5),
	}, func() {
		clay.UI(clay.ID("HeaderIcon"))(clay.ElementDeclaration{
			Image: clay.ImageElementConfig{
				ImageData: cache.GetTexture("assets/textures/ui/hand_cursor.png"),
			},
			BackgroundColor: clay.Color{R: 0.5, G: 0.5, B: 1.0, A: 1.0},
			Layout: clay.LayoutConfig{
				Sizing: clay.Sizing{
					Height: clay.SizingFixed(32),
					Width:  clay.SizingFixed(32),
				},
			},
		}, func() {})
		clay.UI(clay.ID("HeaderLabel"))(clay.ElementDeclaration{
			Layout: clay.LayoutConfig{
				Sizing: clay.Sizing{
					Width:  clay.SizingGrow(0.0),
					Height: clay.SizingFit(0.0, 0.0),
				},
			},
		}, func() {
			clay.OnHover(func(elementId clay.ElementId, pointerInfo clay.PointerData, userData unsafe.Pointer) {
				if pointerInfo.State == clay.POINTER_DATA_PRESSED_THIS_FRAME {
					scr.HeeHeeCount++
				}
			}, unsafe.Pointer(scr))
			clay.Text(strings.Repeat("Hee", scr.HeeHeeCount+1), &clay.TextElementConfig{
				TextColor: clay.Color{R: 1.0, G: 1.0, B: 0.0, A: 1.0},
			})
		})
		clay.UI(clay.ID("HiddenLabel"))(clay.ElementDeclaration{
			Layout: clay.LayoutConfig{
				Sizing: clay.Sizing{
					Width:  clay.SizingFixed(320.0),
					Height: clay.SizingFixed(600.0),
				},
			},
		}, func() {
			clay.Text(
				"Down in the underground you've got to mind the gook.\n"+
					"You've gotta take a scoop and start piling up the poop.\n"+
					"You've gotta work from 9-5 and 5-9 again.\n"+
					"It keeps on coming down the pipe and it never seems to end.\n",
				&clay.TextElementConfig{
					TextColor: clay.Color{R: 1.0, G: 0.0, B: 0.0, A: 1.0},
				},
			)
		})
	})

	sliderPanelElem := clay.ElementDeclaration{
		Layout: clay.LayoutConfig{
			Sizing: clay.Sizing{
				Width:  clay.SizingGrow(0.0),
				Height: clay.SizingFixed(128.0),
			},
			Padding: clay.PaddingAll(16.0),
			ChildAlignment: clay.ChildAlignment{
				X: clay.ALIGN_X_LEFT,
				Y: clay.ALIGN_Y_CENTER,
			},
			LayoutDirection: clay.LEFT_TO_RIGHT,
		},
		BackgroundColor: clay.Color{
			R: 0.0,
			G: 0.0,
			B: 0.5,
			A: 0.5,
		},
	}
	clay.UI()(sliderPanelElem, func() {
		sliderId := clay.ID("slider")
		slider := clay.UI(sliderId)
		sliderDecl := clay.ElementDeclaration{
			Layout: clay.LayoutConfig{
				Sizing: clay.Sizing{
					Width:  clay.SizingGrow(0.0),
					Height: clay.SizingFixed(24.0),
				},
				ChildAlignment: clay.ChildAlignment{
					X: clay.ALIGN_X_LEFT,
					Y: clay.ALIGN_Y_CENTER,
				},
				LayoutDirection: clay.LEFT_TO_RIGHT,
			},
			BackgroundColor: clay.Color{A: 1.0},
		}
		clay.OnHover(scrollBarOnHover, unsafe.Pointer(scr))
		slider(sliderDecl, func() {
			knobElem := clay.UI(clay.ID("knob"))
			decl := clay.ElementDeclaration{
				Layout: clay.LayoutConfig{
					Sizing: clay.Sizing{
						Width:  clay.SizingFixed(16.0),
						Height: clay.SizingFixed(20.0),
					},
				},
				BackgroundColor: clay.Color{R: 0.2, G: 0.5, B: 0.7, A: 1.0},
				Floating: clay.FloatingElementConfig{
					AttachPoints: clay.FloatingAttachPoints{
						Element: clay.ATTACH_POINT_LEFT_TOP,
						Parent:  clay.ATTACH_POINT_LEFT_TOP,
					},
					AttachTo: clay.ATTACH_TO_PARENT,
				},
			}
			parentData := clay.GetElementData(sliderId)
			decl.Floating.Offset = clay.Vector2{
				X: parentData.BoundingBox.Width * scr.SliderValue,
				Y: 0.0,
			}
			if clay.Hovered() {
				decl.BackgroundColor = clay.Color{R: 0.2, G: 0.7, B: 1.0, A: 1.0}
			}
			knobElem(decl, func() {})
		})
	})

	commands := clay.EndLayout()
	for cmd := range commands.Iter() {
		switch cmd.CommandType {
		case clay.RENDER_COMMAND_TYPE_RECTANGLE:
			clayRect := cmd.RenderData.Rectangle
			box := ui.NewBox(ui.Transform{
				Position: mgl32.Vec2{cmd.BoundingBox.X, cmd.BoundingBox.Y},
				Size:     mgl32.Vec2{cmd.BoundingBox.Width, cmd.BoundingBox.Height},
			}, nil)
			box.BgColor = maybe.Some(color.Color{
				R: clayRect.BackgroundColor.R,
				G: clayRect.BackgroundColor.G,
				B: clayRect.BackgroundColor.B,
				A: clayRect.BackgroundColor.A,
			})
			queue.Add(&box)
		case clay.RENDER_COMMAND_TYPE_IMAGE:
			clayImg := cmd.RenderData.Image
			tex, ok := clayImg.ImageData.(*textures.Texture)
			if !ok {
				failure.LogErrWithLocation("Clay image was not passed a texture for its image data")
				return
			}
			box := ui.NewBox(ui.Transform{
				Position: mgl32.Vec2{cmd.BoundingBox.X, cmd.BoundingBox.Y},
				Size:     mgl32.Vec2{cmd.BoundingBox.Width, cmd.BoundingBox.Height},
			}, tex)
			box.BgColor = maybe.Some(color.Color{
				R: clayImg.BackgroundColor.R,
				G: clayImg.BackgroundColor.G,
				B: clayImg.BackgroundColor.B,
				A: clayImg.BackgroundColor.A,
			})
			queue.Add(&box)
		case clay.RENDER_COMMAND_TYPE_TEXT:
			clayText := cmd.RenderData.Text
			text := ui.NewText(ui.Transform{
				Position: mgl32.Vec2{cmd.BoundingBox.X, cmd.BoundingBox.Y},
				Size:     mgl32.Vec2{cmd.BoundingBox.Width, cmd.BoundingBox.Height},
			}, clayText.StringContents.String(), ui.TextConfig{
				Color: maybe.Some(color.Color{
					R: clayText.TextColor.R,
					G: clayText.TextColor.G,
					B: clayText.TextColor.B,
					A: clayText.TextColor.A,
				}),
				WrapWords: false,
			})
			queue.Add(&text)
		case clay.RENDER_COMMAND_TYPE_SCISSOR_START:
			gl.Enable(gl.SCISSOR_TEST)
			gl.Scissor(int32(cmd.BoundingBox.X), int32(cmd.BoundingBox.Y), int32(cmd.BoundingBox.Width), int32(cmd.BoundingBox.Height))
		case clay.RENDER_COMMAND_TYPE_SCISSOR_END:
			gl.Disable(gl.SCISSOR_TEST)
		}
	}
}
